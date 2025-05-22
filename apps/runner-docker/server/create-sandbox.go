package server

import (
	"context"

	pb "github.com/daytonaio/runner/proto"
)

// Sandbox endpoints
func (s *RunnerServer) CreateSandbox(ctx context.Context, req *pb.CreateSandboxRequest) (*pb.CreateSandboxResponse, error) {
	// TODO: Implement sandbox creation logic
	return &pb.CreateSandboxResponse{
		ContainerId: "container-" + req.Id,
	}, nil
}
