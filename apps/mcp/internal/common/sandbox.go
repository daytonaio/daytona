// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/daytonaio/apiclient"
)

func GetSandbox(ctx context.Context, apiClient *apiclient.APIClient, sandboxId *string) (*apiclient.Sandbox, func(), error) {
	var sbx *apiclient.Sandbox

	if sandboxId != nil && *sandboxId != "" {
		sandbox, _, err := apiClient.SandboxAPI.GetSandbox(ctx, *sandboxId).Execute()
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to get sandbox %s: %v", *sandboxId, err)
		}

		if sandbox.State != nil && *sandbox.State != apiclient.SANDBOXSTATE_STARTED {
			_, _, err := apiClient.SandboxAPI.StartSandbox(ctx, *sandboxId).Execute()
			if err != nil {
				return nil, func() {}, fmt.Errorf("failed to start sandbox %s: %v", *sandboxId, err)
			}
		}

		sbx = sandbox
	} else {
		sandbox, _, err := apiClient.SandboxAPI.CreateSandbox(ctx).CreateSandbox(*apiclient.NewCreateSandbox()).Execute()
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to create sandbox: %v", err)
		}

		sbx = sandbox
	}

	stop := func() {
		_, _, err := apiClient.SandboxAPI.StopSandbox(ctx, sbx.Id).Execute()
		if err != nil {
			slog.Warn("Failed to stop sandbox", "sandbox_id", sbx.Id, "error", err)
		}
	}

	return sbx, stop, nil
}
