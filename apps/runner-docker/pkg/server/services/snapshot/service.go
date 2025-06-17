package snapshot

import (
	"io"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/cache"
	"github.com/docker/docker/client"
)

type SnapshotService struct {
	pb.UnimplementedSnapshotServiceServer
	apiClient client.APIClient
	cache     cache.IRunnerCache
	logWriter io.Writer
}

func NewSnapshotService(dockerClient client.APIClient, cache cache.IRunnerCache, logWriter io.Writer) *SnapshotService {
	return &SnapshotService{
		apiClient: dockerClient,
		cache:     cache,
		logWriter: logWriter,
	}
}
