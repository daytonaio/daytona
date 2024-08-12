package tailscale

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

func TestFowardPort(t *testing.T) {

	t.Run("test_forward_port", func(t *testing.T) {
		workspaceId := "daytona_user_1"
		projectName := "project_1"
		targetPort := uint16(5000)

		profile := &config.Profile{
			Id:   "user_1",
			Name: "user_1",
			Api: config.ServerApi{
				Url: "vscode/daytona.io",
				Key: "tryy4444",
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan error, 1)

		go func() {
			hostPort, errChan := ForwardPort(workspaceId, projectName, targetPort, *profile)
			if hostPort == nil {
				done <- fmt.Errorf("hostPort is nil")
				return
			}
			select {
			case err := <-errChan:
				done <- err
			case <-ctx.Done():
				done <- fmt.Errorf("test timed out")
			}
		}()

	})
}
