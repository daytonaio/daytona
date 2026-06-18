// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

// AwaitSnapshotState polls the snapshot until it reaches one of the given
// states. A timeout <= 0 waits indefinitely; on expiry a timeout-category
// error is returned.
func AwaitSnapshotState(ctx context.Context, apiClient *apiclient.APIClient, name string, timeout time.Duration, states ...apiclient.SnapshotState) error {
	var expired <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		expired = timer.C
	}

	for {
		snapshot, res, err := apiClient.SnapshotsAPI.GetSnapshot(ctx, name).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		for _, s := range states {
			if snapshot.State == s {
				return nil
			}
		}

		switch snapshot.State {
		case apiclient.SNAPSHOTSTATE_ERROR, apiclient.SNAPSHOTSTATE_BUILD_FAILED:
			if !snapshot.ErrorReason.IsSet() {
				return fmt.Errorf("snapshot processing failed")
			}
			return fmt.Errorf("snapshot processing failed: %s", *snapshot.ErrorReason.Get())
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-expired:
			return clierr.Newf(clierr.CategoryTimeout, "timed out after %s waiting for snapshot %q to reach state %s", timeout, name, awaitStateNames(states))
		case <-time.After(time.Second):
		}
	}
}

// AwaitSandboxState polls the sandbox until it reaches one of the given
// states. A timeout <= 0 waits indefinitely; on expiry a timeout-category
// error is returned.
func AwaitSandboxState(ctx context.Context, apiClient *apiclient.APIClient, targetSandbox string, timeout time.Duration, states ...apiclient.SandboxState) error {
	var expired <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		expired = timer.C
	}

	for {
		sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, targetSandbox).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		if sandbox.State != nil {
			for _, s := range states {
				if *sandbox.State == s {
					return nil
				}
			}
			if *sandbox.State == apiclient.SANDBOXSTATE_ERROR || *sandbox.State == apiclient.SANDBOXSTATE_BUILD_FAILED {
				if sandbox.ErrorReason == nil {
					return fmt.Errorf("sandbox processing failed")
				}
				return fmt.Errorf("sandbox processing failed: %s", *sandbox.ErrorReason)
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-expired:
			return clierr.Newf(clierr.CategoryTimeout, "timed out after %s waiting for sandbox %q to reach state %s", timeout, targetSandbox, awaitStateNames(states))
		case <-time.After(time.Second):
		}
	}
}

// awaitStateNames renders a list of awaited states for timeout messages.
func awaitStateNames[T ~string](states []T) string {
	names := make([]string, len(states))
	for i, s := range states {
		names[i] = string(s)
	}
	return strings.Join(names, ", ")
}
