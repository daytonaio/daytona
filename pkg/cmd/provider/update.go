// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
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
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		providerList, err := apiclient_util.GetProviderList()
		if err != nil {
			log.Fatal(err)
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		providerManager := manager.NewProviderManager(manager.ProviderManagerConfig{RegistryUrl: *serverConfig.RegistryUrl})

		providersManifest, err := providerManager.GetProvidersManifest()
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

		providerToUpdate, err := provider.GetProviderFromPrompt(providerList, "Choose a provider to update", false)
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

		err = updateProvider(providerToUpdate, providersManifest, apiClient)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Provider %s has been successfully updated\n", *providerToUpdate.Name)
	},
}

func updateProvider(providerToUpdate *apiclient.Provider, providersManifest *manager.ProvidersManifest, apiClient *apiclient.APIClient) error {
	providerManifest, ok := (*providersManifest)[*providerToUpdate.Name]
	if !ok {
		return fmt.Errorf("Provider %s not found in manifest", *providerToUpdate.Name)
	}

	version, ok := providerManifest.Versions["latest"]
	if !ok {
		_, latest := providerManifest.FindLatestVersion()
		version = *latest
	}

	downloadUrls := convertToStringMap(version.DownloadUrls)

	res, err := apiClient.ProviderAPI.InstallProviderExecute(apiclient.ApiInstallProviderRequest{}.Provider(apiclient.InstallProviderRequest{
		Name:         providerToUpdate.Name,
		DownloadUrls: &downloadUrls,
	}))
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	return nil
}

func init() {
	providerUpdateCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Update all providers")
}
