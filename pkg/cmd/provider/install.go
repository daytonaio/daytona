// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/huh"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/provider"
	provider_view "github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/daytonaio/daytona/pkg/views/targetconfig"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var yesFlag bool

var providerInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install provider",
	Args:    cobra.NoArgs,
	Aliases: []string{"i"},
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		providerManager := manager.NewProviderManager(manager.ProviderManagerConfig{RegistryUrl: serverConfig.RegistryUrl})

		providersManifest, err := providerManager.GetProvidersManifest()
		if err != nil {
			return err
		}

		if providersManifest == nil {
			return errors.New("could not get providers manifest")
		}

		providersManifestLatest := providersManifest.GetLatestVersions()
		if providersManifestLatest == nil {
			return errors.New("could not get providers manifest")
		}

		providerList := GetProviderListFromManifest(providersManifestLatest)
		specificProviderName := "Select a specific version"
		specificProviderVersion := ""
		providerList = append(providerList, apiclient.Provider{Name: specificProviderName, Label: &specificProviderName, Version: specificProviderVersion})

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
			providerList = GetProviderListFromManifest(providersManifest)

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

		err = InstallProvider(apiClient, *providerToInstall, providersManifest)
		if err != nil {
			return err
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Provider %s has been successfully installed", providerToInstall.Name))

		targetConfigs, res, err := apiClient.TargetConfigAPI.ListTargetConfigs(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if slices.ContainsFunc(targetConfigs, func(t apiclient.TargetConfig) bool {
			return t.ProviderInfo.Name == providerToInstall.Name
		}) {
			return nil
		}

		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Add a Target Config?").
						Value(&yesFlag),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				return err
			}
		}

		if yesFlag {
			targetConfigManifest, res, err := apiClient.ProviderAPI.GetTargetConfigManifest(context.Background(), providerToInstall.Name).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			targetConfigToSet := &targetconfig.TargetConfigView{
				Options: "{}",
				ProviderInfo: targetconfig.ProviderInfo{
					Name:    providerToInstall.Name,
					Version: providerToInstall.Version,
				},
			}

			err = targetconfig.NewTargetConfigNameInput(&targetConfigToSet.Name, []string{})
			if err != nil {
				return err
			}

			err = targetconfig.SetTargetConfigForm(targetConfigToSet, *targetConfigManifest)
			if err != nil {
				return err
			}

			targetConfigData := apiclient.CreateTargetConfigDTO{
				Name:    targetConfigToSet.Name,
				Options: targetConfigToSet.Options,
				ProviderInfo: apiclient.ProviderProviderInfo{
					Name:    targetConfigToSet.ProviderInfo.Name,
					Version: targetConfigToSet.ProviderInfo.Version,
				},
			}

			res, err = apiClient.TargetConfigAPI.SetTargetConfig(context.Background()).TargetConfig(targetConfigData).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			if err != nil {
				return err
			}

			views.RenderInfoMessage("Target Config set successfully")
		}
		return nil
	},
}

func init() {
	providerInstallCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
}

func GetProviderListFromManifest(manifest *manager.ProvidersManifest) []apiclient.Provider {
	providerList := []apiclient.Provider{}
	for providerName, providerManifest := range *manifest {
		for version := range providerManifest.Versions {
			providerList = append(providerList, apiclient.Provider{
				Name:    providerName,
				Label:   providerManifest.Label,
				Version: version,
			})
		}
	}

	return providerList
}

func ConvertOSToStringMap(downloadUrls map[os.OperatingSystem]string) map[string]string {
	stringMap := map[string]string{}
	for os, url := range downloadUrls {
		stringMap[string(os)] = url
	}

	return stringMap
}

func InstallProvider(apiClient *apiclient.APIClient, providerToInstall provider_view.ProviderView, providersManifest *manager.ProvidersManifest) error {
	downloadUrls := ConvertOSToStringMap((*providersManifest)[providerToInstall.Name].Versions[providerToInstall.Version].DownloadUrls)
	err := views_util.WithInlineSpinner("Installing", func() error {
		res, err := apiClient.ProviderAPI.InstallProviderExecute(apiclient.ApiInstallProviderRequest{}.Provider(apiclient.InstallProviderRequest{
			Name:         providerToInstall.Name,
			DownloadUrls: downloadUrls,
		}))

		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		return nil
	})

	return err
}
