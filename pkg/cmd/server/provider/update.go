// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
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
		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		providerList, err := server.GetProviderList()
		if err != nil {
			log.Fatal(err)
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(serverapiclient.ApiGetConfigRequest{})
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		providersManifest, err := manager.GetProvidersManifest(*serverConfig.RegistryUrl)
		if err != nil {
			log.Fatal(err)
		}

		if allFlag {
			for _, provider := range providerList {
				fmt.Printf("Updating provider %s\n", *provider.Name)
				err := updateProvider(&provider, providersManifest, apiClient)
				if err != nil {
					log.Error(fmt.Sprintf("Failed to update provider %s: %s", *provider.Name, err))
				} else {
					fmt.Printf("Provider %s has been successfully updated\n", *provider.Name)
				}
			}

			return
		}

		providerToUpdate := provider.GetProviderFromPrompt(providerList, "Choose a provider to update")
		if providerToUpdate == nil {
			return
		}

		err = updateProvider(providerToUpdate, providersManifest, apiClient)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Provider %s has been successfully updated\n", *providerToUpdate.Name)
	},
}

func updateProvider(providerToUpdate *serverapiclient.Provider, providersManifest *manager.ProvidersManifest, apiClient *serverapiclient.APIClient) error {
	providerManifest, ok := (*providersManifest)[*providerToUpdate.Name]
	if !ok {
		return fmt.Errorf("Provider %s not found in manifest", *providerToUpdate.Name)
	}

	version, ok := providerManifest.Versions["latest"]
	if !ok {
		version = *manager.FindLatestVersion(providerManifest)
	}

	downloadUrls := convertToStringMap(version.DownloadUrls)

	res, err := apiClient.ProviderAPI.InstallProviderExecute(serverapiclient.ApiInstallProviderRequest{}.Provider(serverapiclient.InstallProviderRequest{
		Name:         providerToUpdate.Name,
		DownloadUrls: &downloadUrls,
	}))
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	return nil
}

func init() {
	providerUpdateCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Update all providers")
}
