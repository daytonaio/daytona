package server

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) PullImage(ctx context.Context, req *pb.PullImageRequest) (*pb.PullImageResponse, error) {
	// TODO: Implement image pull logic
	return &pb.PullImageResponse{
		Message: fmt.Sprintf("Image %s pulled", req.Image),
	}, nil
}
