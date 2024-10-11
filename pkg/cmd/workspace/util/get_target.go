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
	target_view "github.com/daytonaio/daytona/pkg/views/target"
)

func GetTarget(ctx context.Context, apiClient *apiclient.APIClient, targetList []apiclient.ProviderTarget, activeProfileName string, targetNameFlag string) (*target_view.TargetView, error) {
	if targetNameFlag != "" {
		for _, t := range targetList {
			if t.Name == targetNameFlag {
				return util.Pointer(target_view.GetTargetViewFromTarget(t)), nil
			}
		}
		return nil, fmt.Errorf("target '%s' not found", targetNameFlag)
	}

	serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
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

		providerViewList, err = provider.GetProviderViewOptions(apiClient, latestProviders, ctx)
		if err != nil {
			return nil, err
		}
	}

	selectedTarget, err := target_view.GetTargetFromPrompt(targetList, activeProfileName, &providerViewList, false)
	if err != nil {
		return nil, err
	}

	if selectedTarget.ProviderInfo.Installed == nil || *selectedTarget.ProviderInfo.Installed || selectedTarget == nil {
		return selectedTarget, nil
	}

	err = provider.InstallProvider(apiClient, provider_view.ProviderView{
		Name:    selectedTarget.ProviderInfo.Name,
		Version: selectedTarget.ProviderInfo.Version,
	}, providersManifest)
	if err != nil {
		return nil, err
	}

	targetManifest, res, err := apiClient.ProviderAPI.GetTargetManifest(context.Background(), selectedTarget.ProviderInfo.Name).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	selectedTarget.Name = ""
	err = target_view.NewTargetNameInput(&selectedTarget.Name, util.ArrayMap(targetList, func(t apiclient.ProviderTarget) string {
		return t.Name
	}))
	if err != nil {
		return nil, err
	}

	err = target_view.SetTargetForm(selectedTarget, *targetManifest)
	if err != nil {
		return nil, err
	}

	res, err = apiClient.TargetAPI.SetTarget(context.Background()).Target(apiclient.ProviderTarget{
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
