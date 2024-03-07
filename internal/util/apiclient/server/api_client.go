// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"net/http"
	"os"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/types"
)

var apiClient *serverapiclient.APIClient

func GetApiClient(profile *config.Profile) (*serverapiclient.APIClient, error) {
	if apiClient != nil {
		return apiClient, nil
	}

	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	serverUrl := "http://localhost:3000"

	if envApiUrl, ok := os.LookupEnv("DAYTONA_SERVER_API_URL"); ok {
		serverUrl = envApiUrl
	} else {
		var activeProfile config.Profile
		if profile == nil {
			var err error
			activeProfile, err = c.GetActiveProfile()
			if err != nil {
				return nil, err
			}
		} else {
			activeProfile = *profile
		}

		serverUrl = activeProfile.Api.Url
	}

	clientConfig := serverapiclient.NewConfiguration()
	clientConfig.Servers = serverapiclient.ServerConfigurations{
		{
			URL: serverUrl,
		},
	}

	// clientConfig.AddDefaultHeader("Authorization", "Bearer "+token)

	apiClient = serverapiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	return apiClient, nil
}

func ToServerConfig(config *serverapiclient.ServerConfig) *types.ServerConfig {
	return &types.ServerConfig{
		ProvidersDir:      *config.ProvidersDir,
		RegistryUrl:       *config.RegistryUrl,
		Id:                *config.Id,
		ServerDownloadUrl: *config.ServerDownloadUrl,
		Frps: &types.FRPSConfig{
			Domain:   *config.Frps.Domain,
			Port:     uint32(*config.Frps.Port),
			Protocol: *config.Frps.Protocol,
		},
		ApiPort:       uint32(*config.ApiPort),
		HeadscalePort: uint32(*config.HeadscalePort),
	}
}

func GetProviderList() ([]serverapiclient.Provider, error) {
	apiClient, err := GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	providersList, res, err := apiClient.ProviderAPI.ListProviders(ctx).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	return providersList, nil
}

func GetTargetList() ([]serverapiclient.ProviderTarget, error) {
	apiClient, err := GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	targets, resp, err := apiClient.TargetAPI.ListTargets(context.Background()).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(resp, err)
	}

	return targets, nil
}

func GetWorkspace(workspaceNameOrId string) (*serverapiclient.Workspace, error) {
	ctx := context.Background()

	apiClient, err := GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	workspace, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceNameOrId).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	return workspace, nil
}
