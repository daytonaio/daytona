package server

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	// TODO: Implement snapshot creation logic
	return &pb.CreateSnapshotResponse{
		Message: fmt.Sprintf("Snapshot created for workspace %s", req.WorkspaceId),
	}, nil
}
