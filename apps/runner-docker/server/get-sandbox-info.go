package server

import (
	"context"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) GetSandboxInfo(ctx context.Context, req *pb.GetSandboxInfoRequest) (*pb.GetSandboxInfoResponse, error) {
	// TODO: Implement get sandbox info logic
	return &pb.GetSandboxInfoResponse{
		State:         "running",
		SnapshotState: "none",
	}, nil
}
