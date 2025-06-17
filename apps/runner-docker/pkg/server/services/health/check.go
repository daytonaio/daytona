package health

import (
	"context"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
)

func Check(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: pb.HealthStatus_HEALTH_STATUS_HEALTHY}, nil
}
