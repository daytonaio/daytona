// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"slices"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
)

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
