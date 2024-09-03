// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/provider"
	provider_view "github.com/daytonaio/daytona/pkg/views/provider"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var providerInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install provider",
	Args:    cobra.NoArgs,
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
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

		if providersManifest == nil {
			log.Fatal("Could not get providers manifest")
		}

		providersManifestLatest := providersManifest.GetLatestVersions()
		if providersManifestLatest == nil {
			log.Fatal("Could not get providers manifest")
		}

		providerList := GetProviderListFromManifest(providersManifestLatest)
		specificProviderName := "Select a specific version"
		specificProviderVersion := ""
		providerList = append(providerList, apiclient.Provider{Name: specificProviderName, Version: specificProviderVersion})

		providerToInstall, err := provider.GetProviderFromPrompt(provider.ProviderListToView(providerList), "Choose a Provider to Install", false)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return
			} else {
				log.Fatal(err)
			}
		}

		if providerToInstall == nil {
			return
		}

		if providerToInstall.Name == specificProviderName {
			providerList = GetProviderListFromManifest(providersManifest)

			providerToInstall, err = provider.GetProviderFromPrompt(provider.ProviderListToView(providerList), "Choose a specific provider to install", false)
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return
				} else {
					log.Fatal(err)
				}
			}

			if providerToInstall == nil {
				return
			}
		}

		err = InstallProvider(apiClient, *providerToInstall, providersManifest)
		if err != nil {
			log.Fatal(err)
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Provider %s has been successfully installed", providerToInstall.Name))
	},
}

func GetProviderListFromManifest(manifest *manager.ProvidersManifest) []apiclient.Provider {
	pluginList := []apiclient.Provider{}
	for pluginName, pluginManifest := range *manifest {
		for version := range pluginManifest.Versions {
			pluginList = append(pluginList, apiclient.Provider{
				Name:    pluginName,
				Version: version,
			})
		}
	}

	return pluginList
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
