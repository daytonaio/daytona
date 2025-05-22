package server

import (
	"context"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status:  "healthy",
		Version: "1.0.0", // Replace with actual version
	}, nil
}
