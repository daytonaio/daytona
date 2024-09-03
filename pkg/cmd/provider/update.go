// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var allFlag bool

var providerUpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update provider",
	Args:    cobra.NoArgs,
	Aliases: []string{"up"},
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

		serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		providerManager := manager.NewProviderManager(manager.ProviderManagerConfig{RegistryUrl: serverConfig.RegistryUrl})

		providersManifest, err := providerManager.GetProvidersManifest()
		if err != nil {
			log.Fatal(err)
		}

		if allFlag {
			for _, provider := range providerList {
				fmt.Printf("Updating provider %s\n", provider.Name)
				err := updateProvider(provider.Name, providersManifest, apiClient)
				if err != nil {
					log.Error(fmt.Sprintf("Failed to update provider %s: %s", provider.Name, err))
				} else {
					fmt.Printf("Provider %s has been successfully updated\n", provider.Name)
				}
			}

			return
		}

		providerToUpdate, err := provider.GetProviderFromPrompt(provider.ProviderListToView(providerList), "Choose a Provider to Update", false)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return
			} else {
				log.Fatal(err)
			}
		}
		if providerToUpdate == nil {
			return
		}

		err = updateProvider(providerToUpdate.Name, providersManifest, apiClient)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Provider %s has been successfully updated\n", providerToUpdate.Name)
	},
}

func updateProvider(providerName string, providersManifest *manager.ProvidersManifest, apiClient *apiclient.APIClient) error {
	providerManifest, ok := (*providersManifest)[providerName]
	if !ok {
		return fmt.Errorf("Provider %s not found in manifest", providerName)
	}

	version, ok := providerManifest.Versions["latest"]
	if !ok {
		_, latest := providerManifest.FindLatestVersion()
		version = *latest
	}

	downloadUrls := ConvertOSToStringMap(version.DownloadUrls)

	res, err := apiClient.ProviderAPI.InstallProviderExecute(apiclient.ApiInstallProviderRequest{}.Provider(apiclient.InstallProviderRequest{
		Name:         providerName,
		DownloadUrls: downloadUrls,
	}))
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	return nil
}

func init() {
	providerUpdateCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Update all providers")
}
