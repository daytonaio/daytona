// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/runner/providermanager"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/provider"
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

		providerManager := providermanager.GetProviderManager(&providermanager.ProviderManagerConfig{RegistryUrl: serverConfig.RegistryUrl})

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
		return nil
	},
}

func init() {
	providerInstallCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
}

func GetProviderListFromManifest(manifest *providermanager.ProvidersManifest) []apiclient.ProviderInfo {
	providerList := []apiclient.ProviderInfo{}
	for providerName, providerManifest := range *manifest {
		for version := range providerManifest.Versions {
			providerList = append(providerList, apiclient.ProviderInfo{
				Name:    providerName,
				Label:   providerManifest.Label,
				Version: version,
			})
		}
	}

	slices.SortFunc(providerList, func(a, b apiclient.ProviderInfo) int {
		return strings.Compare(a.Name, b.Name)
	})

	return providerList
}

func ConvertOSToStringMap(downloadUrls map[os.OperatingSystem]string) map[string]string {
	stringMap := map[string]string{}
	for os, url := range downloadUrls {
		stringMap[string(os)] = url
	}

	return stringMap
}

func InstallProvider(apiClient *apiclient.APIClient, providerToInstall provider.ProviderView, providersManifest *providermanager.ProvidersManifest) error {
	downloadUrls := ConvertOSToStringMap((*providersManifest)[providerToInstall.Name].Versions[providerToInstall.Version].DownloadUrls)

	err := views_util.WithInlineSpinner("Installing", func() error {

		_, err := apiClient.ProviderAPI.InstallProvider(context.Background(), "TEMP").InstallProviderDto(apiclient.InstallProviderDTO{
			Name:                    providerToInstall.Name,
			ProviderDownloadUrlsDTO: downloadUrls,
		}).Execute()

		// await provider exists in runner body

		return err
	})

	return err
}
