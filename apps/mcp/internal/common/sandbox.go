// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"fmt"

	"github.com/daytonaio/apiclient"

	log "github.com/sirupsen/logrus"
)

func GetSandbox(ctx context.Context, apiClient *apiclient.APIClient, sandboxId *string) (*apiclient.Sandbox, func(), error) {
	stop := func() {
		_, _, err := apiClient.SandboxAPI.StopSandbox(ctx, *sandboxId).Execute()
		if err != nil {
			log.Warnf("failed to stop sandbox %s: %v", *sandboxId, err)
		}
	}

	if sandboxId == nil || *sandboxId == "" {
		sandbox, _, err := apiClient.SandboxAPI.CreateSandbox(ctx).CreateSandbox(*apiclient.NewCreateSandbox()).Execute()
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to create sandbox: %v", err)
		}

		return sandbox, stop, nil
	}

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

	return sandbox, stop, nil
}
