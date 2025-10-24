// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"context"
	"fmt"

	"github.com/daytonaio/apiclient"
)

func GetSandbox(ctx context.Context, apiClient *apiclient.APIClient, sandboxId *string) (*string, error) {
	if sandboxId != nil && *sandboxId != "" {
		sandbox, _, err := apiClient.SandboxAPI.GetSandbox(ctx, *sandboxId).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to get sandbox %s: %v", *sandboxId, err)
		}

		if sandbox.State != nil && *sandbox.State != apiclient.SANDBOXSTATE_STARTED {
			_, _, err := apiClient.SandboxAPI.StartSandbox(ctx, *sandboxId).Execute()
			if err != nil {
				return nil, fmt.Errorf("failed to start sandbox %s: %v", *sandboxId, err)
			}
		}
	} else {
		sandbox, _, err := apiClient.SandboxAPI.CreateSandbox(ctx).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to create sandbox: %v", err)
		}

		sandboxId = &sandbox.Id
	}

	return sandboxId, nil
}
