// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/models"
)

// TODO: review returning nil
func ToTargetConfig(targetConfigDto *apiclient.TargetConfig) *models.TargetConfig {
	if targetConfigDto == nil {
		return nil
	}

	providerInfo := ToProviderInfo(&targetConfigDto.ProviderInfo)
	if providerInfo == nil {
		return nil
	}

	return &models.TargetConfig{
		Id:           targetConfigDto.Id,
		Name:         targetConfigDto.Name,
		ProviderInfo: *providerInfo,
		Options:      targetConfigDto.Options,
		Deleted:      targetConfigDto.Deleted,
	}
}

func ToApiClientTargetConfig(targetConfig *models.TargetConfig) *apiclient.TargetConfig {
	if targetConfig == nil {
		return nil
	}

	providerInfoDto := ToApiClientProviderInfo(&targetConfig.ProviderInfo)
	if providerInfoDto == nil {
		return nil
	}

	return &apiclient.TargetConfig{
		Id:           targetConfig.Id,
		Name:         targetConfig.Name,
		ProviderInfo: *ToApiClientProviderInfo(&targetConfig.ProviderInfo),
		Options:      targetConfig.Options,
		Deleted:      targetConfig.Deleted,
	}
}
