// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
)

type GetTargetConfigParams struct {
	Ctx               context.Context
	ApiClient         *apiclient.APIClient
	ActiveProfileName string
	TargetNameFlag    string
	PromptUsingTUI    bool
}

func GetTarget(params GetTargetConfigParams) (*apiclient.TargetDTO, bool, error) {
	targetList, res, err := params.ApiClient.TargetAPI.ListTargets(params.Ctx).Execute()
	if err != nil {
		return nil, false, apiclient_util.HandleErrorResponse(res, err)
	}

	if params.TargetNameFlag != "" {
		for _, t := range targetList {
			if t.Name == params.TargetNameFlag {
				return &t, false, nil
			}
		}
		return nil, false, fmt.Errorf("target config '%s' not found", params.TargetNameFlag)
	}

	if !params.PromptUsingTUI {
		for _, t := range targetList {
			if t.Default {
				return &t, false, nil
			}
		}
	}

	selectedTarget := selection.GetTargetFromPrompt(targetList, true, "Use")
	if err != nil {
		return nil, false, err
	}

	if selectedTarget == nil {
		return nil, false, nil
	}

	if selectedTarget.Name == selection.NewTargetIdentifier {

	}

	return selectedTarget, true, nil

	// serverConfig, res, err := params.ApiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
	// if err != nil {
	// 	return nil, false, apiclient_util.HandleErrorResponse(res, err)
	// }

	// providersManifest, err := manager.NewProviderManager(manager.ProviderManagerConfig{
	// 	RegistryUrl: serverConfig.RegistryUrl,
	// }).GetProvidersManifest()
	// if err != nil {
	// 	return nil, false, err
	// }

	// var providerViewList []provider_view.ProviderView
	// if providersManifest != nil {
	// 	providersManifestLatest := providersManifest.GetLatestVersions()
	// 	if providersManifestLatest == nil {
	// 		return nil, false, errors.New("could not get latest provider versions")
	// 	}

	// 	latestProviders := provider.GetProviderListFromManifest(providersManifestLatest)

	// 	providerViewList, err = provider.GetProviderViewOptions(params.ApiClient, latestProviders, params.Ctx)
	// 	if err != nil {
	// 		return nil, false, err
	// 	}
	// }

	// selectedTargetConfig, err := targetconfig.GetTargetConfigFromPrompt(params.TargetConfigs, params.ActiveProfileName, &providerViewList, false, "Use")
	// if err != nil {
	// 	return nil, false, err
	// }

	// if selectedTargetConfig.ProviderInfo.Installed == nil || *selectedTargetConfig.ProviderInfo.Installed || selectedTargetConfig == nil {
	// 	return selectedTargetConfig, false, nil
	// }

	// err = provider.InstallProvider(params.ApiClient, provider_view.ProviderView{
	// 	Name:    selectedTargetConfig.ProviderInfo.Name,
	// 	Version: selectedTargetConfig.ProviderInfo.Version,
	// }, providersManifest)
	// if err != nil {
	// 	return nil, err
	// }

	// targetConfigManifest, res, err := params.ApiClient.ProviderAPI.GetTargetConfigManifest(context.Background(), selectedTargetConfig.ProviderInfo.Name).Execute()
	// if err != nil {
	// 	return nil, apiclient_util.HandleErrorResponse(res, err)
	// }

	// selectedTargetConfig.Name = ""
	// err = targetconfig.NewTargetConfigNameInput(&selectedTargetConfig.Name, util.ArrayMap(params.TargetConfigs, func(t apiclient.TargetConfig) string {
	// 	return t.Name
	// }))
	// if err != nil {
	// 	return nil, err
	// }

	// err = targetconfig.SetTargetConfigForm(selectedTargetConfig, *targetConfigManifest)
	// if err != nil {
	// 	return nil, err
	// }

	// res, err = params.ApiClient.TargetConfigAPI.SetTargetConfig(context.Background()).TargetConfig(apiclient.CreateTargetConfigDTO{
	// 	Name:    selectedTargetConfig.Name,
	// 	Options: selectedTargetConfig.Options,
	// 	ProviderInfo: apiclient.TargetProviderInfo{
	// 		Name:    selectedTargetConfig.ProviderInfo.Name,
	// 		Version: selectedTargetConfig.ProviderInfo.Version,
	// 	},
	// }).Execute()
	// if err != nil {
	// 	return nil, apiclient_util.HandleErrorResponse(res, err)
	// }

	// return selectedTargetConfig, nil
}
