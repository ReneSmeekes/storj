// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package metainfo

import (
	"bytes"
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"storj.io/storj/pkg/auth"
	"storj.io/storj/pkg/macaroon"
	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/storj"
	"storj.io/storj/satellite/console"
)

const (
	// BucketNameRestricted feature flag to toggle bucket name validation
	BucketNameRestricted = false

	requestTTL = time.Hour * 4
)

var (
	ipRegexp = regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
)

// TTLItem keeps association between serial number and ttl
type TTLItem struct {
	serialNumber storj.SerialNumber
	ttl          time.Time
}

type createRequest struct {
	Expiration *timestamp.Timestamp
	Redundancy *pb.RedundancyScheme

	ttl time.Time
}

type createRequests struct {
	mu sync.RWMutex
	// orders limit serial number used because with CreateSegment we don't have path yet
	entries map[storj.SerialNumber]*createRequest

	muTTL      sync.Mutex
	entriesTTL []*TTLItem
}

func newCreateRequests() *createRequests {
	return &createRequests{
		entries:    make(map[storj.SerialNumber]*createRequest),
		entriesTTL: make([]*TTLItem, 0),
	}
}

func (requests *createRequests) Put(ctx context.Context, serialNumber storj.SerialNumber, createRequest *createRequest) {
	defer mon.Task()(&ctx)(nil)

	ttl := time.Now().Add(requestTTL)

	go func() {
		requests.muTTL.Lock()
		requests.entriesTTL = append(requests.entriesTTL, &TTLItem{
			serialNumber: serialNumber,
			ttl:          ttl,
		})
		requests.muTTL.Unlock()
	}()

	createRequest.ttl = ttl
	requests.mu.Lock()
	requests.entries[serialNumber] = createRequest
	requests.mu.Unlock()

	go requests.cleanup(ctx)
}

func (requests *createRequests) Load(ctx context.Context, serialNumber storj.SerialNumber) (*createRequest, bool) {
	defer mon.Task()(&ctx)(nil)

	requests.mu.RLock()
	request, found := requests.entries[serialNumber]
	if request != nil && request.ttl.Before(time.Now()) {
		request = nil
		found = false
	}
	requests.mu.RUnlock()

	return request, found
}

func (requests *createRequests) Remove(ctx context.Context, serialNumber storj.SerialNumber) {
	defer mon.Task()(&ctx)(nil)
	requests.mu.Lock()
	delete(requests.entries, serialNumber)
	requests.mu.Unlock()
}

func (requests *createRequests) cleanup(ctx context.Context) {
	defer mon.Task()(&ctx)(nil)

	requests.muTTL.Lock()
	now := time.Now()
	remove := make([]storj.SerialNumber, 0)
	newStart := 0
	for i, item := range requests.entriesTTL {
		if item.ttl.Before(now) {
			remove = append(remove, item.serialNumber)
			newStart = i + 1
		} else {
			break
		}
	}
	requests.entriesTTL = requests.entriesTTL[newStart:]
	requests.muTTL.Unlock()

	for _, serialNumber := range remove {
		requests.Remove(ctx, serialNumber)
	}
}

func (endpoint *Endpoint) validateAuth(ctx context.Context, action macaroon.Action) (_ *console.APIKeyInfo, err error) {
	defer mon.Task()(&ctx)(&err)
	keyData, ok := auth.GetAPIKey(ctx)
	if !ok {
		endpoint.log.Error("unauthorized request", zap.Error(status.Errorf(codes.Unauthenticated, "Invalid API credential")))
		return nil, status.Errorf(codes.Unauthenticated, "Invalid API credential")
	}

	key, err := macaroon.ParseAPIKey(string(keyData))
	if err != nil {
		endpoint.log.Error("unauthorized request", zap.Error(status.Errorf(codes.Unauthenticated, "Invalid API credential")))
		return nil, status.Errorf(codes.Unauthenticated, "Invalid API credential")
	}

	keyInfo, err := endpoint.apiKeys.GetByHead(ctx, key.Head())
	if err != nil {
		endpoint.log.Error("unauthorized request", zap.Error(status.Errorf(codes.Unauthenticated, err.Error())))
		return nil, status.Errorf(codes.Unauthenticated, "Invalid API credential")
	}

	// Revocations are currently handled by just deleting the key.
	err = key.Check(ctx, keyInfo.Secret, action, nil)
	if err != nil {
		endpoint.log.Error("unauthorized request", zap.Error(status.Errorf(codes.Unauthenticated, err.Error())))
		return nil, status.Errorf(codes.Unauthenticated, "Invalid API credential")
	}

	return keyInfo, nil
}

func (endpoint *Endpoint) validateCreateSegment(ctx context.Context, req *pb.SegmentWriteRequest) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = endpoint.validateBucket(ctx, req.Bucket)
	if err != nil {
		return err
	}

	err = endpoint.validateRedundancy(ctx, req.Redundancy)
	if err != nil {
		return err
	}

	return nil
}

func (endpoint *Endpoint) validateCommitSegment(ctx context.Context, req *pb.SegmentCommitRequest) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = endpoint.validateBucket(ctx, req.Bucket)
	if err != nil {
		return err
	}

	err = endpoint.validatePointer(ctx, req.Pointer)
	if err != nil {
		return err
	}

	if req.Pointer.Type == pb.Pointer_REMOTE {
		remote := req.Pointer.Remote

		if len(req.OriginalLimits) == 0 {
			return Error.New("no order limits")
		}
		if int32(len(req.OriginalLimits)) != remote.Redundancy.Total {
			return Error.New("invalid no order limit for piece")
		}

		for _, piece := range remote.RemotePieces {
			limit := req.OriginalLimits[piece.PieceNum]

			err := endpoint.orders.VerifyOrderLimitSignature(ctx, limit)
			if err != nil {
				return err
			}

			if limit == nil {
				return Error.New("invalid no order limit for piece")
			}
			derivedPieceID := remote.RootPieceId.Derive(piece.NodeId)
			if limit.PieceId.IsZero() || limit.PieceId != derivedPieceID {
				return Error.New("invalid order limit piece id")
			}
			if bytes.Compare(piece.NodeId.Bytes(), limit.StorageNodeId.Bytes()) != 0 {
				return Error.New("piece NodeID != order limit NodeID")
			}
		}
	}

	if len(req.OriginalLimits) > 0 {
		createRequest, found := endpoint.createRequests.Load(ctx, req.OriginalLimits[0].SerialNumber)

		switch {
		case !found:
			return Error.New("missing create request or request expired")
		case !proto.Equal(createRequest.Expiration, req.Pointer.ExpirationDate):
			return Error.New("pointer expiration date does not match requested one")
		case !proto.Equal(createRequest.Redundancy, req.Pointer.Remote.Redundancy):
			return Error.New("pointer redundancy scheme date does not match requested one")
		}
	}

	return nil
}

func (endpoint *Endpoint) validateBucket(ctx context.Context, bucket []byte) (err error) {
	defer mon.Task()(&ctx)(&err)

	if len(bucket) == 0 {
		return Error.New("bucket not specified")
	}

	if !BucketNameRestricted {
		return nil
	}

	if len(bucket) < 3 || len(bucket) > 63 {
		return Error.New("bucket name must be at least 3 and no more than 63 characters long")
	}

	// Regexp not used because benchmark shows it will be slower for valid bucket names
	// https://gist.github.com/mniewrzal/49de3af95f36e63e88fac24f565e444c
	labels := bytes.Split(bucket, []byte("."))
	for _, label := range labels {
		err = validateBucketLabel(label)
		if err != nil {
			return err
		}
	}

	if ipRegexp.MatchString(string(bucket)) {
		return Error.New("bucket name cannot be formatted as an IP address")
	}

	return nil
}

func validateBucketLabel(label []byte) error {
	if len(label) == 0 {
		return Error.New("bucket label cannot be empty")
	}

	if !isLowerLetter(label[0]) && !isDigit(label[0]) {
		return Error.New("bucket label must start with a lowercase letter or number")
	}

	if label[0] == '-' || label[len(label)-1] == '-' {
		return Error.New("bucket label cannot start or end with a hyphen")
	}

	for i := 1; i < len(label)-1; i++ {
		if !isLowerLetter(label[i]) && !isDigit(label[i]) && (label[i] != '-') && (label[i] != '.') {
			return Error.New("bucket name must contain only lowercase letters, numbers or hyphens")
		}
	}

	return nil
}

func isLowerLetter(r byte) bool {
	return r >= 'a' && r <= 'z'
}

func isDigit(r byte) bool {
	return r >= '0' && r <= '9'
}

func (endpoint *Endpoint) validatePointer(ctx context.Context, pointer *pb.Pointer) (err error) {
	defer mon.Task()(&ctx)(&err)

	if pointer == nil {
		return Error.New("no pointer specified")
	}

	if pointer.Type == pb.Pointer_INLINE && pointer.Remote != nil {
		return Error.New("pointer type is INLINE but remote segment is set")
	}

	// TODO does it all?
	if pointer.Type == pb.Pointer_REMOTE {
		if pointer.Remote == nil {
			return Error.New("no remote segment specified")
		}
		if pointer.Remote.RemotePieces == nil {
			return Error.New("no remote segment pieces specified")
		}
		if pointer.Remote.Redundancy == nil {
			return Error.New("no redundancy scheme specified")
		}
	}
	return nil
}

func (endpoint *Endpoint) validateRedundancy(ctx context.Context, redundancy *pb.RedundancyScheme) (err error) {
	defer mon.Task()(&ctx)(&err)

	if endpoint.rsConfig.Validate == true {
		if endpoint.rsConfig.ErasureShareSize.Int32() != redundancy.ErasureShareSize ||
			endpoint.rsConfig.MaxThreshold != int(redundancy.Total) ||
			endpoint.rsConfig.MinThreshold != int(redundancy.MinReq) ||
			endpoint.rsConfig.RepairThreshold != int(redundancy.RepairThreshold) ||
			endpoint.rsConfig.SuccessThreshold != int(redundancy.SuccessThreshold) {
			return Error.New("provided redundancy scheme parameters not allowed: want [%d, %d, %d, %d, %d] got [%d, %d, %d, %d, %d]",
				endpoint.rsConfig.MinThreshold,
				endpoint.rsConfig.RepairThreshold,
				endpoint.rsConfig.SuccessThreshold,
				endpoint.rsConfig.MaxThreshold,
				endpoint.rsConfig.ErasureShareSize.Int32(),

				redundancy.MinReq,
				redundancy.RepairThreshold,
				redundancy.SuccessThreshold,
				redundancy.Total,
				redundancy.ErasureShareSize,
			)
		}
	}

	return nil
}
