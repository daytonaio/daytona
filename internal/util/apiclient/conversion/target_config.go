// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs/dto"
)

func ToTargetConfig(createTargetConfigDto dto.CreateTargetConfigDTO) *provider.TargetConfig {
	return &provider.TargetConfig{
		Name:         createTargetConfigDto.Name,
		ProviderInfo: createTargetConfigDto.ProviderInfo,
		Options:      createTargetConfigDto.Options,
	}
}
