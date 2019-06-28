// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package gc

import (
	"context"
	"time"

	"github.com/zeebo/errs"
	"go.uber.org/zap"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"

	"storj.io/storj/pkg/auth/signing"
	"storj.io/storj/pkg/bloomfilter"
	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/storj"
	"storj.io/storj/pkg/transport"
	"storj.io/storj/uplink/piecestore"
)

var (
	mon = monkit.Package()

	// Error defines the garbage queue errors class
	Error = errs.Class("garbage queue error")
)

// RetainInfo contains info needed for a storage node to retain important data and delete garbage data
type RetainInfo struct {
	Filter       *bloomfilter.Filter
	CreationDate time.Time
}

// Queue defines the functions of the satellite's garbage collection "queue"
// TODO: does this need to be an interface?
type Queue interface {
	// Add adds a RetainRequest to the Garbage "queue"
	Add(ctx context.Context, nodeID storj.NodeID, pieceID storj.PieceID) error
	// Send sends the garbage delete requests to all storage nodes
	Send(ctx context.Context) error
}

// Garbage contains a map of storage nodes and their respective garbage delete requests
type Garbage struct {
	log       *zap.Logger
	transport transport.Client
	config    Config
	Requests  map[storj.NodeID]*RetainInfo
}

// NewGarbage instantiates a Garbage "queue"
func NewGarbage(log *zap.Logger, config Config, transport transport.Client) *Garbage {
	return &Garbage{
		log:       log,
		transport: transport,
		config:    config,
		Requests:  make(map[storj.NodeID]*RetainInfo),
	}
}

// Add adds a RetainRequest to the Garbage "queue"
func (Garbage *Garbage) Add(ctx context.Context, nodeID storj.NodeID, pieceID storj.PieceID) (err error) {
	defer mon.Task()(&ctx)(&err)

	var filter *bloomfilter.Filter

	if _, ok := Garbage.Requests[nodeID]; !ok {
		filter = bloomfilter.NewOptimal(int(Garbage.config.InitialPieces), Garbage.config.FalsePositiveRate)
		Garbage.Requests[nodeID].Filter = filter
		Garbage.Requests[nodeID].CreationDate = time.Now().UTC()
	}
	Garbage.Requests[nodeID].Filter.Add(pieceID)
	return nil
}

// Send sends the garbage retain requests to all storage nodes
func (Garbage *Garbage) Send(ctx context.Context) (err error) {
	defer mon.Task()(&ctx)(&err)

	for id := range Garbage.Requests {
		log := Garbage.log.Named(id.String())
		// TODO: access storage node address to populate target
		target := &pb.Node{Id: id}
		signer := signing.SignerFromFullIdentity(Garbage.transport.Identity())

		ps, err := piecestore.Dial(ctx, Garbage.transport, target, log, signer, piecestore.DefaultConfig)
		if err != nil {
			return Error.Wrap(err)
		}
		defer func() {
			err := ps.Close()
			if err != nil {
				Garbage.log.Error("garbage queue failed to close conn to node: %+v", zap.Error(err))
			}
		}()
		// TODO: send the retain request to the storage node
	}

	return nil
}
