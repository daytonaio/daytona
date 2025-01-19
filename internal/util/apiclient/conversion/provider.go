// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"slices"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/services"
)

func GetProviderListFromManifest(manifest *util.ProvidersManifest) []services.ProviderDTO {
	providerList := []services.ProviderDTO{}
	for providerName, providerManifest := range *manifest {
		latestVersion, _ := providerManifest.FindLatestVersion()
		for version := range providerManifest.Versions {
			providerList = append(providerList, services.ProviderDTO{
				Name:    providerName,
				Label:   providerManifest.Label,
				Version: version,
				Latest:  version == latestVersion,
			})
		}
	}

	slices.SortFunc(providerList, func(a, b services.ProviderDTO) int {
		return strings.Compare(a.Name, b.Name)
	})

	return providerList
}
