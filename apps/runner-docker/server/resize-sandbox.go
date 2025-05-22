package server

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) ResizeSandbox(ctx context.Context, req *pb.ResizeSandboxRequest) (*pb.ResizeSandboxResponse, error) {
	// TODO: Implement sandbox resizing logic
	return &pb.ResizeSandboxResponse{
		Message: fmt.Sprintf("Sandbox %s resized", req.WorkspaceId),
	}, nil
}
