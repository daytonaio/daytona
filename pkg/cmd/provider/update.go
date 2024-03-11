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
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

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

		providerToUpdate := provider.GetProviderFromPrompt(providerList, "CHOOSE A PROVIDER TO UPDATE")

		if providerToUpdate == nil {
			return
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(serverapiclient.ApiGetConfigRequest{})
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		providersManifest, err := manager.GetProvidersManifest(*serverConfig.RegistryUrl)
		if err != nil {
			log.Fatal(err)
		}

		if providersManifest == nil {
			log.Fatal("Could not get providers manifest")
		}

		providerManifest, ok := (*providersManifest)[*providerToUpdate.Name]
		if !ok {
			log.Fatal(fmt.Sprintf("Provider %s not found in manifest", *providerToUpdate.Name))
		}

		version, ok := providerManifest.Versions["latest"]
		if !ok {
			version = *manager.FindLatestVersion(providerManifest)
		}

		downloadUrls := convertToStringMap(version.DownloadUrls)

		res, err = apiClient.ProviderAPI.InstallProviderExecute(serverapiclient.ApiInstallProviderRequest{}.Provider(serverapiclient.InstallProviderRequest{
			Name:         providerToUpdate.Name,
			DownloadUrls: &downloadUrls,
		}))
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		if err != nil {
			log.Fatal(err)
		}

		view_util.RenderInfoMessageBold(fmt.Sprintf("Provider %s has been successfully updated", *providerToUpdate.Name))
	},
}
