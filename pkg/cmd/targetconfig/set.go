// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"context"
	"errors"
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
	"github.com/daytonaio/daytona/pkg/views/targetconfig"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var TargetConfigSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set target config",
	Args:    cobra.NoArgs,
	Aliases: []string{"s", "add", "update", "register", "edit"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		targetConfig, err := TargetConfigCreationFlow(ctx, apiClient, activeProfile.Name, true)
		if err != nil {
			return err
		}

		views.RenderInfoMessage(fmt.Sprintf("Target config '%s' set successfully", targetConfig.Name))
		return nil
	},
}

func TargetConfigCreationFlow(ctx context.Context, apiClient *apiclient.APIClient, activeProfileName string, allowUpdating bool) (*targetconfig.TargetConfigView, error) {
	var isNewProvider bool

	serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
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
			return nil, errors.New("could not get latest provider versions")
		}

		latestProviders = provider.GetProviderListFromManifest(providersManifestLatest)
	} else {
		fmt.Println("Could not get provider manifest. Can't check for new providers to install")
	}

	providerViewList, err := provider.GetProviderViewOptions(apiClient, latestProviders, ctx)
	if err != nil {
		return nil, err
	}

	selectedProvider, err := provider_view.GetProviderFromPrompt(providerViewList, "Choose a Provider", false)
	if err != nil {
		if common.IsCtrlCAbort(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if selectedProvider == nil {
		return nil, nil
	}

	if selectedProvider.Installed != nil && !*selectedProvider.Installed {
		if providersManifest == nil {
			return nil, errors.New("could not get providers manifest")
		}
		err = provider.InstallProvider(apiClient, *selectedProvider, providersManifest)
		if err != nil {
			return nil, err
		}
		isNewProvider = true
	}

	targetConfigs, res, err := apiClient.TargetConfigAPI.ListTargetConfigs(ctx).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	filteredConfigs := []apiclient.TargetConfig{}
	for _, t := range targetConfigs {
		if t.ProviderInfo.Name == selectedProvider.Name {
			filteredConfigs = append(filteredConfigs, t)
		}
	}

	var selectedTargetConfig *targetconfig.TargetConfigView

	if !isNewProvider || len(filteredConfigs) > 0 {
		selectedTargetConfig, err = targetconfig.GetTargetConfigFromPrompt(filteredConfigs, activeProfileName, nil, true, "Set")
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return nil, nil
			} else {
				return nil, err
			}
		}
	} else {
		selectedTargetConfig = &targetconfig.TargetConfigView{
			Name:    targetconfig.NewTargetConfigName,
			Options: "{}",
		}
	}

	if selectedTargetConfig.Name == targetconfig.NewTargetConfigName {
		selectedTargetConfig.Name = ""
		err = targetconfig.NewTargetConfigNameInput(&selectedTargetConfig.Name, internal_util.ArrayMap(targetConfigs, func(t apiclient.TargetConfig) string {
			return t.Name
		}))
		if err != nil {
			return nil, err
		}
	} else {
		if !allowUpdating {
			return selectedTargetConfig, nil
		}
	}

	targetConfigManifest, res, err := apiClient.ProviderAPI.GetTargetConfigManifest(context.Background(), selectedProvider.Name).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	err = targetconfig.SetTargetConfigForm(selectedTargetConfig, *targetConfigManifest)
	if err != nil {
		return nil, err
	}

	targetConfigData := apiclient.CreateTargetConfigDTO{
		Name:    selectedTargetConfig.Name,
		Options: selectedTargetConfig.Options,
		ProviderInfo: apiclient.TargetProviderInfo{
			Name:    selectedProvider.Name,
			Version: selectedProvider.Version,
		},
	}

	// TODO: consider returning the DTO from the api
	res, err = apiClient.TargetConfigAPI.SetTargetConfig(context.Background()).TargetConfig(targetConfigData).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	return &targetconfig.TargetConfigView{
		Name:    selectedTargetConfig.Name,
		Options: selectedTargetConfig.Options,
		ProviderInfo: targetconfig.ProviderInfo{
			Name:    selectedProvider.Name,
			Version: selectedProvider.Version,
		},
	}, nil
}
