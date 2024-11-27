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
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var allFlag bool

var providerUpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update provider",
	Args:    cobra.NoArgs,
	Aliases: []string{"up"},
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

		if len(providerList) == 0 {
			views_util.NotifyEmptyProviderList(true)
			return nil
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		providerManager := manager.GetProviderManager(&manager.ProviderManagerConfig{RegistryUrl: serverConfig.RegistryUrl})

		providersManifest, err := providerManager.GetProvidersManifest()
		if err != nil {
			return err
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

			return nil
		}

		providerToUpdate, err := provider.GetProviderFromPrompt(provider.ProviderListToView(providerList), "Choose a Provider to Update", false)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return nil
			} else {
				return err
			}
		}
		if providerToUpdate == nil {
			return nil
		}

		err = updateProvider(providerToUpdate.Name, providersManifest, apiClient)
		if err != nil {
			return err
		}

		fmt.Printf("Provider %s has been successfully updated\n", providerToUpdate.Name)
		return nil
	},
}

func updateProvider(providerName string, providersManifest *manager.ProvidersManifest, apiClient *apiclient.APIClient) error {
	providerManifest, ok := (*providersManifest)[providerName]
	if !ok {
		return fmt.Errorf("provider %s not found in manifest", providerName)
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
