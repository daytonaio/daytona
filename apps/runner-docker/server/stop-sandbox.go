package server

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) StopSandbox(ctx context.Context, req *pb.StopSandboxRequest) (*pb.StopSandboxResponse, error) {
	// TODO: Implement sandbox stop logic
	return &pb.StopSandboxResponse{
		Message: fmt.Sprintf("Sandbox %s stopped", req.WorkspaceId),
	}, nil
}
