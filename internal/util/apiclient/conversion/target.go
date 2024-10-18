// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/providertargets/dto"
)

func ToProviderTarget(createProviderTargetDto dto.CreateProviderTargetDTO) *provider.ProviderTarget {
	return &provider.ProviderTarget{
		Name:         createProviderTargetDto.Name,
		ProviderInfo: createProviderTargetDto.ProviderInfo,
		Options:      createProviderTargetDto.Options,
	}
}
