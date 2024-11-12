// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs/dto"
)

func ToTargetConfig(createTargetConfigDto dto.CreateTargetConfigDTO) *models.TargetConfig {
	return &models.TargetConfig{
		Name:         createTargetConfigDto.Name,
		ProviderInfo: createTargetConfigDto.ProviderInfo,
		Options:      createTargetConfigDto.Options,
	}
}
