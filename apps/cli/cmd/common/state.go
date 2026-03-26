// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"fmt"
	"time"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func AwaitSnapshotState(ctx context.Context, apiClient *apiclient.APIClient, name string, state apiclient.SnapshotState) error {
	for {
		snapshot, res, err := apiClient.SnapshotsAPI.GetSnapshot(ctx, name).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		switch snapshot.State {
		case state:
			return nil
		case apiclient.SNAPSHOTSTATE_ERROR, apiclient.SNAPSHOTSTATE_BUILD_FAILED:
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
		sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, targetSandbox).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		if sandbox.State != nil && *sandbox.State == state {
			return nil
		} else if sandbox.State != nil && (*sandbox.State == apiclient.SANDBOXSTATE_ERROR || *sandbox.State == apiclient.SANDBOXSTATE_BUILD_FAILED) {
			if sandbox.ErrorReason == nil {
				return fmt.Errorf("sandbox processing failed")
			}
			return fmt.Errorf("sandbox processing failed: %s", *sandbox.ErrorReason)
		}

		time.Sleep(time.Second)
	}
}
