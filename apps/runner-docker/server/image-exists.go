package server

import (
	"context"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) ImageExists(ctx context.Context, req *pb.ImageExistsRequest) (*pb.ImageExistsResponse, error) {
	// TODO: Implement image exists check logic
	return &pb.ImageExistsResponse{
		Exists: true,
	}, nil
}
