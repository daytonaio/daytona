/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"fmt"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/runner/pkg/common"
)

type snapshotSandboxJobResult struct {
	Ref        string   `json:"ref"`
	Hash       string   `json:"hash"`
	SizeGB     float64  `json:"sizeGB"`
	Entrypoint []string `json:"entrypoint,omitempty"`
	Cmd        []string `json:"cmd,omitempty"`
}

func (e *Executor) snapshotSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	var payload SnapshotSandboxPayload
	if err := e.parsePayload(job.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot sandbox payload: %w", err)
	}

	if payload.Registry == nil {
		return nil, fmt.Errorf("registry is required for sandbox snapshot")
	}

	containerID := job.GetResourceId()
	if containerID == "" {
		return nil, fmt.Errorf("job resource id (sandbox id) is required")
	}

	info, err := e.docker.CreateSnapshotFromSandbox(ctx, containerID, payload.Registry)
	if err != nil {
		return nil, common.FormatRecoverableError(err)
	}

	return snapshotSandboxJobResult{
		Ref:        info.Name,
		Hash:       info.Hash,
		SizeGB:     info.SizeGB,
		Entrypoint: info.Entrypoint,
		Cmd:        info.Cmd,
	}, nil
}
