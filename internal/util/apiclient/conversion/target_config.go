// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/server/targetconfigs/dto"
	"github.com/daytonaio/daytona/pkg/target/config"
)

func ToTargetConfig(createTargetConfigDto dto.CreateTargetConfigDTO) *config.TargetConfig {
	return &config.TargetConfig{
		Name:         createTargetConfigDto.Name,
		ProviderInfo: createTargetConfigDto.ProviderInfo,
		Options:      createTargetConfigDto.Options,
	}
}
