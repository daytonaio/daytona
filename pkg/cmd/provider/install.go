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
	"github.com/daytonaio/daytona/pkg/views/target"
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

		targets, res, err := apiClient.TargetAPI.ListTargets(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if slices.ContainsFunc(targets, func(t apiclient.ProviderTarget) bool {
			return t.ProviderInfo.Name == providerToInstall.Name
		}) {
			return nil
		}

		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Add a Target?").
						Value(&yesFlag),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				return err
			}
		}

		if yesFlag {
			targetManifest, res, err := apiClient.ProviderAPI.GetTargetManifest(context.Background(), providerToInstall.Name).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			targetToSet := &target.TargetView{
				Options: "{}",
				ProviderInfo: target.ProviderInfo{
					Name:    providerToInstall.Name,
					Version: providerToInstall.Version,
				},
			}

			err = target.NewTargetNameInput(&targetToSet.Name, []string{})
			if err != nil {
				return err
			}

			err = target.SetTargetForm(targetToSet, *targetManifest)
			if err != nil {
				return err
			}

			targetData := apiclient.ProviderTarget{
				Name:    targetToSet.Name,
				Options: targetToSet.Options,
				ProviderInfo: apiclient.ProviderProviderInfo{
					Name:    targetToSet.ProviderInfo.Name,
					Version: targetToSet.ProviderInfo.Version,
				},
			}

			res, err = apiClient.TargetAPI.SetTarget(context.Background()).Target(targetData).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			if err != nil {
				return err
			}

			views.RenderInfoMessage("Target set successfully")
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
				Label:   &providerManifest.Label,
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
	err := views_util.WithSpinner("Installing", func() error {
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
