// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/spf13/cobra"
)

var providerListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List installed providers",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		providerList, res, err := apiClient.ProviderAPI.ListProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(providerList)
			formattedData.Print()
			return nil
		}

		provider.List(providerList)
		return nil
	},
}

func GetProviderViewOptions(ctx context.Context, apiClient *apiclient.APIClient, latestProviders []apiclient.ProviderInfo) ([]provider.ProviderView, error) {
	var result []provider.ProviderView

	installedProviders, res, err := apiClient.ProviderAPI.ListProviders(ctx).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	providerMap := make(map[string]provider.ProviderView)

	for _, installedProvider := range installedProviders {
		providerMap[installedProvider.Name] = provider.ProviderView{
			Name:                 installedProvider.Name,
			Label:                installedProvider.Label,
			Version:              installedProvider.Version,
			Installed:            util.Pointer(true),
			RunnerId:             installedProvider.RunnerId,
			RunnerName:           installedProvider.RunnerName,
			TargetConfigManifest: installedProvider.TargetConfigManifest,
		}
	}

	for _, latestProvider := range latestProviders {
		if _, exists := providerMap[latestProvider.Name]; !exists {
			providerMap[latestProvider.Name] = provider.ProviderView{
				Name:                 latestProvider.Name,
				Label:                latestProvider.Label,
				Version:              latestProvider.Version,
				Installed:            util.Pointer(false),
				RunnerId:             latestProvider.RunnerId,
				RunnerName:           latestProvider.RunnerName,
				TargetConfigManifest: latestProvider.TargetConfigManifest,
			}
		}
	}

	for _, provider := range providerMap {
		result = append(result, provider)
	}

	return result, nil
}

func init() {
	format.RegisterFormatFlag(providerListCmd)
}
