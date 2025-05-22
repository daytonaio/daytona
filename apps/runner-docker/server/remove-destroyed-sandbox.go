package server

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) RemoveDestroyedSandbox(ctx context.Context, req *pb.RemoveDestroyedSandboxRequest) (*pb.RemoveDestroyedSandboxResponse, error) {
	// TODO: Implement remove destroyed sandbox logic
	return &pb.RemoveDestroyedSandboxResponse{
		Message: fmt.Sprintf("Destroyed sandbox %s removed", req.WorkspaceId),
	}, nil
}
