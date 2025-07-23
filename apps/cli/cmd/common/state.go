// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
)

func AwaitSnapshotState(ctx context.Context, apiClient *apiclient.APIClient, name string, state apiclient.SnapshotState) error {
	for {
		snapshot, res, err := apiClient.SnapshotsAPI.GetSnapshot(ctx, name).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		if snapshot.State == state {
			return nil
		} else if snapshot.State == apiclient.SNAPSHOTSTATE_ERROR || snapshot.State == apiclient.SNAPSHOTSTATE_BUILD_FAILED {
			if !snapshot.ErrorReason.IsSet() {
				return fmt.Errorf("snapshot processing failed")
			}
			return fmt.Errorf("snapshot processing failed: %s", *snapshot.ErrorReason.Get())
		}

		time.Sleep(time.Second)
	}
}

func AwaitSandboxState(ctx context.Context, apiClient *apiclient.APIClient, targetSandbox string, state apiclient.SandboxState) error {
	for {
		sandboxes, res, err := apiClient.SandboxAPI.ListSandboxes(ctx).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		for _, sandbox := range sandboxes {
			if sandbox.Id == targetSandbox {
				if sandbox.State != nil && *sandbox.State == state {
					return nil
				} else if sandbox.State != nil && (*sandbox.State == apiclient.SANDBOXSTATE_ERROR || *sandbox.State == apiclient.SANDBOXSTATE_BUILD_FAILED) {
					if sandbox.ErrorReason == nil {
						return fmt.Errorf("sandbox processing failed")
					}
					return fmt.Errorf("sandbox processing failed: %s", *sandbox.ErrorReason)
				}
			}
		}

		time.Sleep(time.Second)
	}
}
