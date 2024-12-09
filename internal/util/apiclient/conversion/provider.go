// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"slices"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/models"
)

func ToApiClientProviderInfo(provider *models.ProviderInfo) *apiclient.ProviderInfo {
	targetConfigManifest := make(map[string]apiclient.TargetConfigProperty)

	for key, value := range provider.TargetConfigManifest {
		targetConfigManifest[key] = apiclient.TargetConfigProperty{
			DefaultValue:      &value.DefaultValue,
			Description:       &value.Description,
			DisabledPredicate: &value.DisabledPredicate,
			InputMasked:       &value.InputMasked,
			Options:           value.Options,
			Suggestions:       value.Suggestions,
			Type:              util.Pointer(apiclient.ModelsTargetConfigPropertyType(value.Type)),
		}
	}

	return &apiclient.ProviderInfo{
		AgentlessTarget:      &provider.AgentlessTarget,
		Label:                provider.Label,
		Name:                 provider.Name,
		RunnerId:             provider.RunnerId,
		Version:              provider.Version,
		TargetConfigManifest: targetConfigManifest,
	}
}

func ToProviderInfo(providerDto *apiclient.ProviderInfo) *models.ProviderInfo {
	targetConfigManifest := make(map[string]models.TargetConfigProperty)

	for key, value := range providerDto.TargetConfigManifest {
		targetConfigManifest[key] = models.TargetConfigProperty{
			DefaultValue:      *value.DefaultValue,
			Description:       *value.Description,
			DisabledPredicate: *value.DisabledPredicate,
			InputMasked:       *value.InputMasked,
			Options:           value.Options,
			Suggestions:       value.Suggestions,
			Type:              models.TargetConfigPropertyType(*value.Type),
		}
	}

	result := &models.ProviderInfo{
		Label:                providerDto.Label,
		Name:                 providerDto.Name,
		RunnerId:             providerDto.RunnerId,
		Version:              providerDto.Version,
		TargetConfigManifest: targetConfigManifest,
	}

	if providerDto.AgentlessTarget != nil {
		result.AgentlessTarget = *providerDto.AgentlessTarget
	}

	return result
}

func GetProviderListFromManifest(manifest *util.ProvidersManifest) []apiclient.ProviderInfo {
	providerList := []apiclient.ProviderInfo{}
	for providerName, providerManifest := range *manifest {
		for version := range providerManifest.Versions {
			providerList = append(providerList, apiclient.ProviderInfo{
				Name:    providerName,
				Label:   providerManifest.Label,
				Version: version,
			})
		}
	}

	slices.SortFunc(providerList, func(a, b apiclient.ProviderInfo) int {
		return strings.Compare(a.Name, b.Name)
	})

	return providerList
}
