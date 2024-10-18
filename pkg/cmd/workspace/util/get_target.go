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
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	provider_view "github.com/daytonaio/daytona/pkg/views/provider"
	target_view "github.com/daytonaio/daytona/pkg/views/target"
)

type GetTargetConfig struct {
	Ctx               context.Context
	ApiClient         *apiclient.APIClient
	TargetList        []apiclient.ProviderTarget
	ActiveProfileName string
	TargetNameFlag    string
	PromptUsingTUI    bool
}

func GetTarget(config GetTargetConfig) (*target_view.TargetView, error) {
	if config.TargetNameFlag != "" {
		for _, t := range config.TargetList {
			if t.Name == config.TargetNameFlag {
				return util.Pointer(target_view.GetTargetViewFromTarget(t)), nil
			}
		}
		return nil, fmt.Errorf("target '%s' not found", config.TargetNameFlag)
	}

	if !config.PromptUsingTUI {
		for _, t := range config.TargetList {
			if t.IsDefault {
				return util.Pointer(target_view.GetTargetViewFromTarget(t)), nil
			}
		}
	}

	serverConfig, res, err := config.ApiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
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

		providerViewList, err = provider.GetProviderViewOptions(config.ApiClient, latestProviders, config.Ctx)
		if err != nil {
			return nil, err
		}
	}

	selectedTarget, err := target_view.GetTargetFromPrompt(config.TargetList, config.ActiveProfileName, &providerViewList, false, "Use")
	if err != nil {
		if common.IsCtrlCAbort(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if selectedTarget.ProviderInfo.Installed == nil || *selectedTarget.ProviderInfo.Installed || selectedTarget == nil {
		return selectedTarget, nil
	}

	err = provider.InstallProvider(config.ApiClient, provider_view.ProviderView{
		Name:    selectedTarget.ProviderInfo.Name,
		Version: selectedTarget.ProviderInfo.Version,
	}, providersManifest)
	if err != nil {
		return nil, err
	}

	targetManifest, res, err := config.ApiClient.ProviderAPI.GetTargetManifest(context.Background(), selectedTarget.ProviderInfo.Name).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	selectedTarget.Name = ""
	err = target_view.NewTargetNameInput(&selectedTarget.Name, util.ArrayMap(config.TargetList, func(t apiclient.ProviderTarget) string {
		return t.Name
	}))
	if err != nil {
		return nil, err
	}

	err = target_view.SetTargetForm(selectedTarget, *targetManifest)
	if err != nil {
		return nil, err
	}

	res, err = config.ApiClient.TargetAPI.SetTarget(context.Background()).Target(apiclient.CreateProviderTargetDTO{
		Name:    selectedTarget.Name,
		Options: selectedTarget.Options,
		ProviderInfo: apiclient.ProviderProviderInfo{
			Name:    selectedTarget.ProviderInfo.Name,
			Version: selectedTarget.ProviderInfo.Version,
		},
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	return selectedTarget, nil
}
