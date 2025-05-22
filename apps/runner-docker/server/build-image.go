package server

import (
	"context"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) BuildImage(ctx context.Context, req *pb.BuildImageRequest) (*pb.BuildImageResponse, error) {
	// TODO: Implement image build logic
	return &pb.BuildImageResponse{
		Message: "Image built successfully",
	}, nil
}
