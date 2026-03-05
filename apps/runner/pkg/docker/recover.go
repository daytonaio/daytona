// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models"
)

func (d *DockerClient) RecoverSandbox(ctx context.Context, sandboxId string, recoverDto dto.RecoverSandboxDTO) error {
	// Deduce recovery type from error reason, falling back to backup error reason
	recoveryType := common.DeduceRecoveryType(recoverDto.ErrorReason)
	if recoveryType == models.UnknownRecoveryType {
		recoveryType = common.DeduceRecoveryType(recoverDto.BackupErrorReason)
	}
	if recoveryType == models.UnknownRecoveryType {
		return fmt.Errorf("unable to deduce recovery type from error reason: %s, backup error reason: %s", recoverDto.ErrorReason, recoverDto.BackupErrorReason)
	}

	switch recoveryType {
	case models.RecoveryTypeStorageExpansion:
		return d.RecoverFromStorageLimit(ctx, sandboxId, float64(recoverDto.StorageQuota), recoverDto)
	default:
		return fmt.Errorf("unsupported recovery type: %s", recoveryType)
	}
}
