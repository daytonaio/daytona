package server

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) DestroySandbox(ctx context.Context, req *pb.DestroySandboxRequest) (*pb.DestroySandboxResponse, error) {
	// TODO: Implement sandbox destruction logic
	return &pb.DestroySandboxResponse{
		Message: fmt.Sprintf("Sandbox %s destroyed", req.WorkspaceId),
	}, nil
}
