// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/provider"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	provider_view "github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/daytonaio/daytona/pkg/views/targetconfig"
)

type GetTargetConfigParams struct {
	Ctx                  context.Context
	ApiClient            *apiclient.APIClient
	TargetConfigs        []apiclient.TargetConfig
	ActiveProfileName    string
	TargetConfigNameFlag string
	PromptUsingTUI       bool
}

func GetTargetConfig(params GetTargetConfigParams) (*targetconfig.TargetConfigView, error) {
	if params.TargetConfigNameFlag != "" {
		for _, t := range params.TargetConfigs {
			if t.Name == params.TargetConfigNameFlag {
				return util.Pointer(targetconfig.ToTargetConfigView(t)), nil
			}
		}
		return nil, fmt.Errorf("target config '%s' not found", params.TargetConfigNameFlag)
	}

	if !params.PromptUsingTUI {
		for _, t := range params.TargetConfigs {
			if t.IsDefault {
				return util.Pointer(targetconfig.ToTargetConfigView(t)), nil
			}
		}
	}

	serverConfig, res, err := params.ApiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	providersManifest, err := manager.NewProviderManager(manager.ProviderManagerConfig{
		RegistryUrl: serverConfig.RegistryUrl,
	}).GetProvidersManifest()
	if err != nil {
		return nil, err
	}

	var providerViewList []provider_view.ProviderView
	if providersManifest != nil {
		providersManifestLatest := providersManifest.GetLatestVersions()
		if providersManifestLatest == nil {
			return nil, errors.New("could not get latest provider versions")
		}

		latestProviders := provider.GetProviderListFromManifest(providersManifestLatest)

		providerViewList, err = provider.GetProviderViewOptions(params.ApiClient, latestProviders, params.Ctx)
		if err != nil {
			return nil, err
		}
	}

	selectedTargetConfig, err := targetconfig.GetTargetConfigFromPrompt(params.TargetConfigs, params.ActiveProfileName, &providerViewList, false, "Use")
	if err != nil {
		return nil, err
	}

	if selectedTargetConfig.ProviderInfo.Installed == nil || *selectedTargetConfig.ProviderInfo.Installed || selectedTargetConfig == nil {
		return selectedTargetConfig, nil
	}

	err = provider.InstallProvider(params.ApiClient, provider_view.ProviderView{
		Name:    selectedTargetConfig.ProviderInfo.Name,
		Version: selectedTargetConfig.ProviderInfo.Version,
	}, providersManifest)
	if err != nil {
		return nil, err
	}

	targetConfigManifest, res, err := params.ApiClient.ProviderAPI.GetTargetConfigManifest(context.Background(), selectedTargetConfig.ProviderInfo.Name).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	selectedTargetConfig.Name = ""
	err = targetconfig.NewTargetConfigNameInput(&selectedTargetConfig.Name, util.ArrayMap(params.TargetConfigs, func(t apiclient.TargetConfig) string {
		return t.Name
	}))
	if err != nil {
		return nil, err
	}

	err = targetconfig.SetTargetConfigForm(selectedTargetConfig, *targetConfigManifest)
	if err != nil {
		return nil, err
	}

	res, err = params.ApiClient.TargetConfigAPI.SetTargetConfig(context.Background()).TargetConfig(apiclient.CreateTargetConfigDTO{
		Name:    selectedTargetConfig.Name,
		Options: selectedTargetConfig.Options,
		ProviderInfo: apiclient.ProviderProviderInfo{
			Name:    selectedTargetConfig.ProviderInfo.Name,
			Version: selectedTargetConfig.ProviderInfo.Version,
		},
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	return selectedTargetConfig, nil
}
