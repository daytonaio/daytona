// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	internal_util "github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/provider"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/views"
	provider_view "github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/daytonaio/daytona/pkg/views/target"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var TargetSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set provider target",
	Args:    cobra.NoArgs,
	Aliases: []string{"s", "add", "update", "register", "edit"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		providersManifest, err := manager.NewProviderManager(manager.ProviderManagerConfig{
			RegistryUrl: serverConfig.RegistryUrl,
		}).GetProvidersManifest()
		if err != nil {
			log.Error(err)
		}

		var latestProviders []apiclient.Provider
		if providersManifest != nil {
			providersManifestLatest := providersManifest.GetLatestVersions()
			if providersManifestLatest == nil {
				log.Fatal("Could not get latest provider versions")
			}

			latestProviders = provider.GetProviderListFromManifest(providersManifestLatest)
		} else {
			fmt.Println("Could not get provider manifest. Can't check for new providers to install")
		}

		providerViewList := provider.GetProviderViewOptions(apiClient, latestProviders, ctx)

		selectedProvider, err := provider_view.GetProviderFromPrompt(providerViewList, "Choose a Provider", false)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return
			} else {
				log.Fatal(err)
			}
		}

		if selectedProvider == nil {
			return
		}

		if selectedProvider.Installed != nil && !*selectedProvider.Installed {
			if providersManifest == nil {
				log.Fatal("Could not get providers manifest")
			}
			err = provider.InstallProvider(apiClient, *selectedProvider, providersManifest)
			if err != nil {
				log.Fatal(err)
			}
		}

		targets, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		filteredTargets := []apiclient.ProviderTarget{}
		for _, t := range targets {
			if t.ProviderInfo.Name == selectedProvider.Name {
				filteredTargets = append(filteredTargets, t)
			}
		}

		selectedTarget, err := target.GetTargetFromPrompt(filteredTargets, activeProfile.Name, true)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return
			} else {
				log.Fatal(err)
			}
		}

		targetManifest, res, err := apiClient.ProviderAPI.GetTargetManifest(context.Background(), selectedProvider.Name).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		if selectedTarget.Name == target.NewTargetName {
			selectedTarget.Name = ""
			err = target.NewTargetNameInput(&selectedTarget.Name, internal_util.ArrayMap(targets, func(t apiclient.ProviderTarget) string {
				return t.Name
			}))
			if err != nil {
				log.Fatal(err)
			}
		}

		err = target.SetTargetForm(selectedTarget, *targetManifest)
		if err != nil {
			log.Fatal(err)
		}

		selectedTarget.ProviderInfo = apiclient.ProviderProviderInfo{
			Name:    selectedProvider.Name,
			Version: selectedProvider.Version,
		}

		res, err = apiClient.TargetAPI.SetTarget(context.Background()).Target(*selectedTarget).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage("Target set successfully")
	},
}
