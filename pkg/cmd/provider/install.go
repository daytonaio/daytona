// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/views"
	provider_install "github.com/daytonaio/daytona/pkg/views/provider/install"
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

			if selectedRunner == nil {
				return nil
			}

			selectedRunnerId = selectedRunner.Id
		} else {
			selectedRunnerId = args[0]
		}

		availableProviderList, res, err := apiClient.ProviderAPI.ListProvidersForInstall(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		var latestProviderList []apiclient.ProviderDTO
		for _, provider := range availableProviderList {
			if provider.Latest {
				latestProviderList = append(latestProviderList, provider)
			}
		}

		specificProviderName := "Select a specific version"
		latestProviderList = append(latestProviderList, apiclient.ProviderDTO{
			Name:    specificProviderName,
			Label:   &specificProviderName,
			Version: "",
			Latest:  false,
		})

		providerToInstall, err := provider_install.GetProviderFromInstallPrompt(provider_install.ProviderInstallListToView(latestProviderList), "Choose a Provider to Install", false)
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
			providerToInstall, err = provider_install.GetProviderFromInstallPrompt(provider_install.ProviderInstallListToView(availableProviderList), "Choose a specific provider to install", false)
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

		err = InstallProvider(apiClient, selectedRunnerId, *providerToInstall)
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

func InstallProvider(apiClient *apiclient.APIClient, runnerId string, providerToInstall provider_install.ProviderInstallView) error {
	err := views_util.WithInlineSpinner("Installing", func() error {
		res, err := apiClient.ProviderAPI.InstallProvider(context.Background(), runnerId).InstallProviderDto(apiclient.InstallProviderDTO{
			Name:    providerToInstall.Name,
			Version: providerToInstall.Version,
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
