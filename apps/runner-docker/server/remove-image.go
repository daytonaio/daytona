package server

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) RemoveImage(ctx context.Context, req *pb.RemoveImageRequest) (*pb.RemoveImageResponse, error) {
	// TODO: Implement image removal logic
	return &pb.RemoveImageResponse{
		Message: fmt.Sprintf("Image %s removed", req.Image),
	}, nil
}
