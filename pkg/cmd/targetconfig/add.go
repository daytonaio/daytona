// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"context"
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	internal_util "github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/cmd/provider"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	provider_view "github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/daytonaio/daytona/pkg/views/targetconfig"
	"github.com/spf13/cobra"
)

var TargetConfigAddCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add target config",
	Args:    cobra.NoArgs,
	Aliases: []string{"a", "set", "register", "new", "create"},
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

		targetConfig, err := TargetConfigCreationFlow(ctx, apiClient, activeProfile.Name)
		if err != nil {
			return err
		}

		if targetConfig == nil {
			return nil
		}

		views.RenderInfoMessage(fmt.Sprintf("Target config '%s' set successfully", targetConfig.Name))
		return nil
	},
}

func TargetConfigCreationFlow(ctx context.Context, apiClient *apiclient.APIClient, activeProfileName string) (*targetconfig.TargetConfigView, error) {
	serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	providersManifest, err := util.GetProvidersManifest(serverConfig.RegistryUrl)
	if err != nil {
		return nil, err
	}

	var latestProviders []apiclient.ProviderInfo
	if providersManifest != nil {
		providersManifestLatest := providersManifest.GetLatestVersions()
		if providersManifestLatest == nil {
			return nil, errors.New("could not get latest provider versions")
		}

		latestProviders = conversion.GetProviderListFromManifest(providersManifestLatest)
	} else {
		fmt.Println("Could not get provider manifest. Can't check for new providers to install")
	}

	providerViewList, err := provider.GetProviderViewOptions(ctx, apiClient, latestProviders)
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

	selectedTargetConfig := &targetconfig.TargetConfigView{
		Name:    "",
		Options: "{}",
		ProviderInfo: targetconfig.ProviderInfo{
			Name:       selectedProvider.Name,
			RunnerId:   selectedProvider.RunnerId,
			RunnerName: selectedProvider.RunnerName,
			Version:    selectedProvider.Version,
			Label:      selectedProvider.Label,
		},
	}

	if selectedProvider.Installed != nil && !*selectedProvider.Installed {
		if providersManifest == nil {
			return nil, errors.New("could not get providers manifest")
		}

		selectedRunner, err := cmd_common.GetRunnerFlow(apiClient, "Manage Providers")
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return nil, nil
			} else {
				return nil, err
			}
		}

		if selectedRunner == nil {
			return nil, nil
		}

		err = provider.InstallProvider(apiClient, selectedRunner.Id, *selectedProvider, providersManifest)
		if err != nil {
			return nil, err
		}

		selectedProvider.RunnerId = selectedRunner.Id
		selectedProvider.RunnerName = selectedRunner.Name
	}

	targetConfigs, res, err := apiClient.TargetConfigAPI.ListTargetConfigs(ctx).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}
	selectedTargetConfig.Name = ""
	err = targetconfig.NewTargetConfigNameInput(&selectedTargetConfig.Name, internal_util.ArrayMap(targetConfigs, func(t apiclient.TargetConfig) string {
		return t.Name
	}))
	if err != nil {
		return nil, err
	}

	err = targetconfig.SetTargetConfigForm(selectedTargetConfig, selectedProvider.TargetConfigManifest)
	if err != nil {
		return nil, err
	}

	targetConfigData := apiclient.AddTargetConfigDTO{
		Name:    selectedTargetConfig.Name,
		Options: selectedTargetConfig.Options,
		ProviderInfo: apiclient.ProviderInfo{
			Name:                 selectedProvider.Name,
			Version:              selectedProvider.Version,
			Label:                selectedProvider.Label,
			RunnerId:             selectedProvider.RunnerId,
			RunnerName:           selectedProvider.RunnerName,
			TargetConfigManifest: selectedProvider.TargetConfigManifest,
		},
	}

	targetConfig, res, err := apiClient.TargetConfigAPI.AddTargetConfig(context.Background()).TargetConfig(targetConfigData).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	return &targetconfig.TargetConfigView{
		Id:      targetConfig.Id,
		Name:    targetConfig.Name,
		Options: targetConfig.Options,
		ProviderInfo: targetconfig.ProviderInfo{
			Name:       targetConfig.ProviderInfo.Name,
			RunnerId:   targetConfig.ProviderInfo.RunnerId,
			RunnerName: targetConfig.ProviderInfo.RunnerName,
			Version:    targetConfig.ProviderInfo.Version,
			Label:      targetConfig.ProviderInfo.Label,
		},
	}, nil
}
