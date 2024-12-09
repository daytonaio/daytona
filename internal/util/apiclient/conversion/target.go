// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/models"
)

// TODO: review - missing properties
func ToTarget(targetDto *apiclient.TargetDTO) *models.Target {
	if targetDto == nil {
		return nil
	}

	target := &models.Target{
		Id:             targetDto.Id,
		Name:           targetDto.Name,
		TargetConfigId: targetDto.TargetConfigId,
		TargetConfig: models.TargetConfig{
			Id:           targetDto.TargetConfigId,
			Name:         targetDto.TargetConfig.Name,
			ProviderInfo: *ToProviderInfo(&targetDto.TargetConfig.ProviderInfo),
			Options:      targetDto.TargetConfig.Options,
			Deleted:      targetDto.TargetConfig.Deleted,
		},
		EnvVars:          targetDto.EnvVars,
		IsDefault:        targetDto.Default,
		ProviderMetadata: targetDto.ProviderMetadata,
	}

	return target
}
