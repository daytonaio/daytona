package s3

import (
	"github.com/daytonaio/runner/pkg/volume"
)

func init() {
	// Register S3 provider factory
	volume.RegisterProviderFactory("s3", func() volume.Provider {
		return NewS3Provider()
	})
}
