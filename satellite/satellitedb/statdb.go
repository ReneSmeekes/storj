// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package satellitedb

import (
	"context"
	"database/sql"
	"strings"

	"github.com/zeebo/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"

	pb "storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/statdb"
	"storj.io/storj/pkg/storj"
	dbx "storj.io/storj/satellite/satellitedb/dbx"
)

var (
	mon             = monkit.Package()
	errAuditSuccess = errs.Class("statdb audit success error")
	errUptime       = errs.Class("statdb uptime error")
)

// StatDB implements the statdb RPC service
type statDB struct {
	db *dbx.DB
}

// Create a db entry for the provided storagenode
func (s *statDB) Create(ctx context.Context, createReq *statdb.CreateRequest) (resp *statdb.CreateResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var (
		totalAuditCount    int64
		auditSuccessCount  int64
		auditSuccessRatio  float64
		totalUptimeCount   int64
		uptimeSuccessCount int64
		uptimeRatio        float64
	)

	stats := createReq.Stats
	if stats != nil {
		totalAuditCount = stats.AuditCount
		auditSuccessCount = stats.AuditSuccessCount
		auditSuccessRatio, err = checkRatioVars(auditSuccessCount, totalAuditCount)
		if err != nil {
			return nil, errAuditSuccess.Wrap(err)
		}

		totalUptimeCount = stats.UptimeCount
		uptimeSuccessCount = stats.UptimeSuccessCount
		uptimeRatio, err = checkRatioVars(uptimeSuccessCount, totalUptimeCount)
		if err != nil {
			return nil, errUptime.Wrap(err)
		}
	}

	node := createReq.Node

	dbNode, err := s.db.Create_Node(
		ctx,
		dbx.Node_Id(node.Bytes()),
		dbx.Node_AuditSuccessCount(auditSuccessCount),
		dbx.Node_TotalAuditCount(totalAuditCount),
		dbx.Node_AuditSuccessRatio(auditSuccessRatio),
		dbx.Node_UptimeSuccessCount(uptimeSuccessCount),
		dbx.Node_TotalUptimeCount(totalUptimeCount),
		dbx.Node_UptimeRatio(uptimeRatio),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	nodeStats := &pb.NodeStats{
		NodeId:            node,
		AuditSuccessRatio: dbNode.AuditSuccessRatio,
		AuditCount:        dbNode.TotalAuditCount,
		UptimeRatio:       dbNode.UptimeRatio,
		UptimeCount:       dbNode.TotalUptimeCount,
	}
	return &statdb.CreateResponse{
		Stats: nodeStats,
	}, nil
}

// Get a storagenode's stats from the db
func (s *statDB) Get(ctx context.Context, getReq *statdb.GetRequest) (resp *statdb.GetResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	dbNode, err := s.db.Get_Node_By_Id(ctx, dbx.Node_Id(getReq.NodeId.Bytes()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	nodeStats := &pb.NodeStats{
		NodeId:            getReq.NodeId,
		AuditSuccessRatio: dbNode.AuditSuccessRatio,
		AuditCount:        dbNode.TotalAuditCount,
		UptimeRatio:       dbNode.UptimeRatio,
		UptimeCount:       dbNode.TotalUptimeCount,
	}
	return &statdb.GetResponse{
		Stats: nodeStats,
	}, nil
}

// FindInvalidNodes finds a subset of storagenodes that fail to meet minimum reputation requirements
func (s *statDB) FindInvalidNodes(ctx context.Context, getReq *statdb.FindInvalidNodesRequest) (resp *statdb.FindInvalidNodesResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var invalidIds storj.NodeIDList

	nodeIds := getReq.NodeIds
	maxAuditSuccess := getReq.MaxStats.AuditSuccessRatio
	maxUptime := getReq.MaxStats.UptimeRatio

	rows, err := s.findInvalidNodesQuery(nodeIds, maxAuditSuccess, maxUptime)

	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			//@TODO ASK s.log.Error(err.Error())
		}
	}()

	for rows.Next() {
		node := &dbx.Node{}
		err = rows.Scan(&node.Id, &node.TotalAuditCount, &node.TotalUptimeCount, &node.AuditSuccessRatio, &node.UptimeRatio)
		if err != nil {
			return nil, err
		}
		id, err := storj.NodeIDFromBytes(node.Id)
		if err != nil {
			return nil, err
		}
		invalidIds = append(invalidIds, id)
	}

	return &statdb.FindInvalidNodesResponse{
		InvalidIds: invalidIds,
	}, nil
}

func (s *statDB) findInvalidNodesQuery(nodeIds storj.NodeIDList, auditSuccess, uptime float64) (*sql.Rows, error) {
	args := make([]interface{}, len(nodeIds))
	for i, id := range nodeIds {
		args[i] = id.Bytes()
	}
	args = append(args, auditSuccess, uptime)

	rows, err := s.db.Query(`SELECT nodes.id, nodes.total_audit_count,
		nodes.total_uptime_count, nodes.audit_success_ratio,
		nodes.uptime_ratio
		FROM nodes
		WHERE nodes.id IN (?`+strings.Repeat(", ?", len(nodeIds)-1)+`)
		AND nodes.total_audit_count > 0
		AND nodes.total_uptime_count > 0
		AND (
			nodes.audit_success_ratio < ?
			OR nodes.uptime_ratio < ?
		)`, args...)

	return rows, err
}

// Update a single storagenode's stats in the db
func (s *statDB) Update(ctx context.Context, updateReq *statdb.UpdateRequest) (resp *statdb.UpdateResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	createIfReq := &statdb.CreateEntryIfNotExistsRequest{
		Node: updateReq.Node,
	}

	_, err = s.CreateEntryIfNotExists(ctx, createIfReq)
	if err != nil {
		return nil, err
	}

	dbNode, err := s.db.Get_Node_By_Id(ctx, dbx.Node_Id(updateReq.Node.Bytes()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	auditSuccessCount := dbNode.AuditSuccessCount
	totalAuditCount := dbNode.TotalAuditCount
	var auditSuccessRatio float64
	uptimeSuccessCount := dbNode.UptimeSuccessCount
	totalUptimeCount := dbNode.TotalUptimeCount
	var uptimeRatio float64

	updateFields := dbx.Node_Update_Fields{}

	if updateReq.UpdateAuditSuccess {
		auditSuccessCount, totalAuditCount, auditSuccessRatio = updateRatioVars(
			updateReq.AuditSuccess,
			auditSuccessCount,
			totalAuditCount,
		)

		updateFields.AuditSuccessCount = dbx.Node_AuditSuccessCount(auditSuccessCount)
		updateFields.TotalAuditCount = dbx.Node_TotalAuditCount(totalAuditCount)
		updateFields.AuditSuccessRatio = dbx.Node_AuditSuccessRatio(auditSuccessRatio)
	}
	if updateReq.UpdateUptime {
		uptimeSuccessCount, totalUptimeCount, uptimeRatio = updateRatioVars(
			updateReq.IsUp,
			uptimeSuccessCount,
			totalUptimeCount,
		)

		updateFields.UptimeSuccessCount = dbx.Node_UptimeSuccessCount(uptimeSuccessCount)
		updateFields.TotalUptimeCount = dbx.Node_TotalUptimeCount(totalUptimeCount)
		updateFields.UptimeRatio = dbx.Node_UptimeRatio(uptimeRatio)
	}

	dbNode, err = s.db.Update_Node_By_Id(ctx, dbx.Node_Id(updateReq.Node.Bytes()), updateFields)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	nodeStats := &pb.NodeStats{
		NodeId:            updateReq.Node,
		AuditSuccessRatio: dbNode.AuditSuccessRatio,
		AuditCount:        dbNode.TotalAuditCount,
		UptimeRatio:       dbNode.UptimeRatio,
		UptimeCount:       dbNode.TotalUptimeCount,
	}
	return &statdb.UpdateResponse{
		Stats: nodeStats,
	}, nil
}

// UpdateUptime updates a single storagenode's uptime stats in the db
func (s *statDB) UpdateUptime(ctx context.Context, updateReq *statdb.UpdateUptimeRequest) (resp *statdb.UpdateUptimeResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	node := updateReq.GetNode()

	dbNode, err := s.db.Get_Node_By_Id(ctx, dbx.Node_Id(node.Id.Bytes()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	uptimeSuccessCount := dbNode.UptimeSuccessCount
	totalUptimeCount := dbNode.TotalUptimeCount
	var uptimeRatio float64

	updateFields := dbx.Node_Update_Fields{}

	uptimeSuccessCount, totalUptimeCount, uptimeRatio = updateRatioVars(
		node.IsUp,
		uptimeSuccessCount,
		totalUptimeCount,
	)

	updateFields.UptimeSuccessCount = dbx.Node_UptimeSuccessCount(uptimeSuccessCount)
	updateFields.TotalUptimeCount = dbx.Node_TotalUptimeCount(totalUptimeCount)
	updateFields.UptimeRatio = dbx.Node_UptimeRatio(uptimeRatio)

	dbNode, err = s.db.Update_Node_By_Id(ctx, dbx.Node_Id(node.Id.Bytes()), updateFields)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	nodeStats := &pb.NodeStats{
		NodeId:            node.Id,
		AuditSuccessRatio: dbNode.AuditSuccessRatio,
		AuditCount:        dbNode.TotalAuditCount,
		UptimeRatio:       dbNode.UptimeRatio,
		UptimeCount:       dbNode.TotalUptimeCount,
	}
	return &statdb.UpdateUptimeResponse{
		Stats: nodeStats,
	}, nil
}

// UpdateAuditSuccess updates a single storagenode's uptime stats in the db
func (s *statDB) UpdateAuditSuccess(ctx context.Context, updateReq *statdb.UpdateAuditSuccessRequest) (resp *statdb.UpdateAuditSuccessResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	node := updateReq.GetNode()

	dbNode, err := s.db.Get_Node_By_Id(ctx, dbx.Node_Id(node.Id.Bytes()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	auditSuccessCount := dbNode.AuditSuccessCount
	totalAuditCount := dbNode.TotalAuditCount
	var auditRatio float64

	updateFields := dbx.Node_Update_Fields{}

	auditSuccessCount, totalAuditCount, auditRatio = updateRatioVars(
		node.AuditSuccess,
		auditSuccessCount,
		totalAuditCount,
	)

	updateFields.AuditSuccessCount = dbx.Node_AuditSuccessCount(auditSuccessCount)
	updateFields.TotalAuditCount = dbx.Node_TotalAuditCount(totalAuditCount)
	updateFields.AuditSuccessRatio = dbx.Node_AuditSuccessRatio(auditRatio)

	dbNode, err = s.db.Update_Node_By_Id(ctx, dbx.Node_Id(node.Id.Bytes()), updateFields)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	nodeStats := &pb.NodeStats{
		NodeId:            node.Id,
		AuditSuccessRatio: dbNode.AuditSuccessRatio,
		AuditCount:        dbNode.TotalAuditCount,
		UptimeRatio:       dbNode.UptimeRatio,
		UptimeCount:       dbNode.TotalUptimeCount,
	}
	return &statdb.UpdateAuditSuccessResponse{
		Stats: nodeStats,
	}, nil
}

// UpdateBatch for updating multiple farmers' stats in the db
func (s *statDB) UpdateBatch(ctx context.Context, updateBatchReq *statdb.UpdateBatchRequest) (resp *statdb.UpdateBatchResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var nodeStatsList []*pb.NodeStats
	var failedNodes []*statdb.UpdateRequest
	for _, node := range updateBatchReq.NodeList {
		updateReq := &statdb.UpdateRequest{
			Node:               node.Node,
			UpdateAuditSuccess: node.UpdateAuditSuccess,
			AuditSuccess:       node.AuditSuccess,
			UpdateUptime:       node.UpdateUptime,
			IsUp:               node.IsUp,
		}

		updateRes, err := s.Update(ctx, updateReq)
		if err != nil {
			//@TODO ASK s.log.Error(err.Error())
			failedNodes = append(failedNodes, node)
		} else {
			nodeStatsList = append(nodeStatsList, updateRes.Stats)
		}
	}

	updateBatchRes := &statdb.UpdateBatchResponse{
		FailedNodes: failedNodes,
		StatsList:   nodeStatsList,
	}
	return updateBatchRes, nil
}

// CreateEntryIfNotExists creates a statdb node entry and saves to statdb if it didn't already exist
func (s *statDB) CreateEntryIfNotExists(ctx context.Context, createIfReq *statdb.CreateEntryIfNotExistsRequest) (resp *statdb.CreateEntryIfNotExistsResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	getReq := &statdb.GetRequest{
		NodeId: createIfReq.Node,
	}
	getRes, err := s.Get(ctx, getReq)
	// TODO: figure out better way to confirm error is type dbx.ErrorCode_NoRows
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {
		createReq := &statdb.CreateRequest{
			Node: createIfReq.Node,
		}
		res, err := s.Create(ctx, createReq)
		if err != nil {
			return nil, err
		}
		createEntryIfNotExistsRes := &statdb.CreateEntryIfNotExistsResponse{
			Stats: res.Stats,
		}
		return createEntryIfNotExistsRes, nil
	}
	if err != nil {
		return nil, err
	}
	createEntryIfNotExistsRes := &statdb.CreateEntryIfNotExistsResponse{
		Stats: getRes.Stats,
	}
	return createEntryIfNotExistsRes, nil
}

func updateRatioVars(newStatus bool, successCount, totalCount int64) (int64, int64, float64) {
	totalCount++
	if newStatus {
		successCount++
	}
	newRatio := float64(successCount) / float64(totalCount)
	return successCount, totalCount, newRatio
}

func checkRatioVars(successCount, totalCount int64) (ratio float64, err error) {
	if successCount < 0 {
		return 0, errs.New("success count less than 0")
	}
	if totalCount < 0 {
		return 0, errs.New("total count less than 0")
	}
	if successCount > totalCount {
		return 0, errs.New("success count greater than total count")
	}

	ratio = float64(successCount) / float64(totalCount)
	return ratio, nil
}
