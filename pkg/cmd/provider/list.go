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
	provider_view "github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var providerListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List installed providers",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		providerList, res, err := apiClient.ProviderAPI.ListProviders(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(providerList)
			formattedData.Print()
			return
		}

		provider.List(providerList)
	},
}

func GetProviderViewOptions(apiClient *apiclient.APIClient, latestProviders []apiclient.Provider, ctx context.Context) []provider_view.ProviderView {
	var result []provider_view.ProviderView

	installedProviders, res, err := apiClient.ProviderAPI.ListProviders(ctx).Execute()
	if err != nil {
		log.Fatal(apiclient_util.HandleErrorResponse(res, err))
	}

	providerMap := make(map[string]provider_view.ProviderView)

	for _, installedProvider := range installedProviders {
		providerMap[installedProvider.Name] = provider_view.ProviderView{
			Name:      installedProvider.Name,
			Version:   installedProvider.Version,
			Installed: util.Pointer(true),
		}
	}

	for _, latestProvider := range latestProviders {
		if _, exists := providerMap[latestProvider.Name]; !exists {
			providerMap[latestProvider.Name] = provider_view.ProviderView{
				Name:      latestProvider.Name,
				Version:   latestProvider.Version,
				Installed: util.Pointer(false),
			}
		}
	}

	for _, provider := range providerMap {
		result = append(result, provider)
	}

	return result
}

func init() {
	format.RegisterFormatFlag(providerListCmd)
}
