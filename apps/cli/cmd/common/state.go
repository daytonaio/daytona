// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/cli/apiclient"
	daytonaapiclient "github.com/daytonaio/daytona/daytonaapiclient"
)

func AwaitSnapshotState(ctx context.Context, apiClient *daytonaapiclient.APIClient, targetImage string, state daytonaapiclient.SnapshotState) error {
	for {
		snapshots, res, err := apiClient.SnapshotsAPI.GetAllSnapshots(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		for _, snapshot := range snapshots.Items {
			if snapshot.Name == targetImage {
				if snapshot.State == state {
					return nil
				} else if snapshot.State == daytonaapiclient.SNAPSHOTSTATE_ERROR || snapshot.State == daytonaapiclient.SNAPSHOTSTATE_BUILD_FAILED {
					if !snapshot.ErrorReason.IsSet() {
						return fmt.Errorf("snapshot processing failed")
					}
					return fmt.Errorf("snapshot processing failed: %s", *snapshot.ErrorReason.Get())
				}
			}
		}

		time.Sleep(time.Second)
	}
}

func AwaitSandboxState(ctx context.Context, apiClient *daytonaapiclient.APIClient, targetSandbox string, state daytonaapiclient.SandboxState) error {
	for {
		sandboxes, res, err := apiClient.SandboxAPI.ListSandboxes(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		for _, sandbox := range sandboxes {
			if sandbox.Id == targetSandbox {
				if sandbox.State != nil && *sandbox.State == state {
					return nil
				} else if sandbox.State != nil && (*sandbox.State == daytonaapiclient.SANDBOXSTATE_ERROR || *sandbox.State == daytonaapiclient.SANDBOXSTATE_BUILD_FAILED) {
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
