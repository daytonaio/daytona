package server

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) StartSandbox(ctx context.Context, req *pb.StartSandboxRequest) (*pb.StartSandboxResponse, error) {
	// TODO: Implement sandbox start logic
	return &pb.StartSandboxResponse{
		Message: fmt.Sprintf("Sandbox %s started", req.WorkspaceId),
	}, nil
}
