// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/provider"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var yesFlag bool

var installCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install provider",
	Args:    cobra.MaximumNArgs(1),
	Aliases: cmd_common.GetAliases("install"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedRunnerId string

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

			if selectedRunner == nil {
				return nil
			}

			selectedRunnerId = selectedRunner.Id
		} else {
			selectedRunnerId = args[0]
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		providersManifest, err := util.GetProvidersManifest(serverConfig.RegistryUrl)
		if err != nil {
			return err
		}

		providersManifestLatest := providersManifest.GetLatestVersions()
		if providersManifestLatest == nil {
			return errors.New("could not get providers manifest")
		}

		providerList := conversion.GetProviderListFromManifest(providersManifestLatest)
		specificProviderName := "Select a specific version"
		specificProviderVersion := ""
		providerList = append(providerList, apiclient.ProviderInfo{Name: specificProviderName, Label: &specificProviderName, Version: specificProviderVersion})

		providerToInstall, err := provider.GetProviderFromPrompt(provider.ProviderListToView(providerList), "Choose a Provider to Install", false)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return nil
			} else {
				return err
			}
		}

		if providerToInstall == nil {
			return nil
		}

		if providerToInstall.Name == specificProviderName {
			providerList = conversion.GetProviderListFromManifest(providersManifest)

			providerToInstall, err = provider.GetProviderFromPrompt(provider.ProviderListToView(providerList), "Choose a specific provider to install", false)
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}

			if providerToInstall == nil {
				return nil
			}
		}

		err = InstallProvider(apiClient, selectedRunnerId, *providerToInstall, providersManifest)
		if err != nil {
			return err
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Provider %s has been successfully installed", providerToInstall.Name))
		return nil
	},
}

func init() {
	installCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
}

func InstallProvider(apiClient *apiclient.APIClient, runnerId string, providerToInstall provider.ProviderView, providersManifest *util.ProvidersManifest) error {
	downloadUrls := convertOSToStringMap((*providersManifest)[providerToInstall.Name].Versions[providerToInstall.Version].DownloadUrls)

	err := views_util.WithInlineSpinner("Installing", func() error {
		res, err := apiClient.ProviderAPI.InstallProvider(context.Background(), runnerId).InstallProviderDto(apiclient.InstallProviderDTO{
			Name:         providerToInstall.Name,
			DownloadUrls: downloadUrls,
			Version:      providerToInstall.Version,
		}).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		return cmd_common.AwaitProviderInstalled(runnerId, providerToInstall.Name, providerToInstall.Version)
	})

	return err
}

func convertOSToStringMap(downloadUrls map[os.OperatingSystem]string) map[string]string {
	stringMap := map[string]string{}
	for os, url := range downloadUrls {
		stringMap[string(os)] = url
	}

	return stringMap
}
