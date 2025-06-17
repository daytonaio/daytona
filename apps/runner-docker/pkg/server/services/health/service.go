package health

import (
	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
)

type HealthService struct {
	pb.UnimplementedHealthServiceServer
}

// new service
func NewHealthService() *HealthService {
	return &HealthService{}
}
