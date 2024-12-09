// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/common"
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
		var selectedRunnerId string

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			selectedRunner, err := cmd_common.GetRunnerFlow(apiClient, "Manage Providers")
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}

			selectedRunnerId = selectedRunner.Id
		} else {
			selectedRunnerId = args[0]
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

		providersManifest, err := util.GetProvidersManifest(serverConfig.RegistryUrl)
		if err != nil {
			return err
		}

		if allFlag {
			for _, provider := range providerList {
				fmt.Printf("Updating provider %s\n", provider.Name)
				err := updateProvider(selectedRunnerId, provider.Name, providersManifest, apiClient)
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

		err = updateProvider(selectedRunnerId, providerToUpdate.Name, providersManifest, apiClient)
		if err != nil {
			return err
		}

		fmt.Printf("Provider %s has been successfully updated\n", providerToUpdate.Name)
		return nil
	},
}

func updateProvider(runnerId string, providerName string, providersManifest *util.ProvidersManifest, apiClient *apiclient.APIClient) error {
	providerManifest, ok := (*providersManifest)[providerName]
	if !ok {
		return fmt.Errorf("provider %s not found in manifest", providerName)
	}

	version, ok := providerManifest.Versions["latest"]
	if !ok {
		_, latest := providerManifest.FindLatestVersion()
		version = *latest
	}

	downloadUrls := convertOSToStringMap(version.DownloadUrls)

	res, err := apiClient.ProviderAPI.UpdateProvider(context.Background(), runnerId, providerName).DownloadUrls(downloadUrls).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	return nil
}

func init() {
	providerUpdateCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Update all providers")
}
