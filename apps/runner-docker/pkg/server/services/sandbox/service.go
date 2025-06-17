package sandbox

import (
	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
)

type SandboxService struct {
	pb.UnimplementedSandboxServiceServer
}

func NewSandboxService() *SandboxService {
	return &SandboxService{}
}
