// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/constants"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/server"
)

const CLIENT_VERSION_HEADER = "X-Client-Version"

var apiClient *apiclient.APIClient

func GetApiClient(profile *config.Profile) (*apiclient.APIClient, error) {
	if apiClient != nil {
		return apiClient, nil
	}

	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

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

	serverUrl := activeProfile.Api.Url
	apiKey := activeProfile.Api.Key

	healthUrl, err := url.JoinPath(serverUrl, constants.HEALTH_CHECK_ROUTE)
	if err != nil {
		return nil, err
	}

	_, err = http.Head(healthUrl)
	if err != nil {
		return nil, ErrHealthCheckFailed(healthUrl)
	}

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: serverUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	clientConfig.AddDefaultHeader(CLIENT_VERSION_HEADER, internal.Version)

	apiClient = apiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	return apiClient, nil
}

func GetAgentApiClient(apiUrl, apiKey string) (*apiclient.APIClient, error) {
	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: apiUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	clientConfig.AddDefaultHeader(CLIENT_VERSION_HEADER, internal.Version)

	apiClient = apiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	return apiClient, nil
}

func ToServerConfig(config apiclient.ServerConfig) server.Config {
	return server.Config{
		ProvidersDir:      *config.ProvidersDir,
		RegistryUrl:       *config.RegistryUrl,
		Id:                *config.Id,
		ServerDownloadUrl: *config.ServerDownloadUrl,
		Frps: &server.FRPSConfig{
			Domain:   *config.Frps.Domain,
			Port:     uint32(*config.Frps.Port),
			Protocol: *config.Frps.Protocol,
		},
		ApiPort:       uint32(*config.ApiPort),
		HeadscalePort: uint32(*config.HeadscalePort),
	}
}

func GetProviderList() ([]apiclient.Provider, error) {
	apiClient, err := GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	providersList, res, err := apiClient.ProviderAPI.ListProviders(ctx).Execute()
	if err != nil {
		return nil, HandleErrorResponse(res, err)
	}

	return providersList, nil
}

func GetTargetList() ([]apiclient.ProviderTarget, error) {
	apiClient, err := GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	targets, resp, err := apiClient.TargetAPI.ListTargets(context.Background()).Execute()
	if err != nil {
		return nil, HandleErrorResponse(resp, err)
	}

	return targets, nil
}

func GetWorkspace(workspaceNameOrId string) (*apiclient.WorkspaceDTO, error) {
	ctx := context.Background()

	apiClient, err := GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	workspace, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceNameOrId).Execute()
	if err != nil {
		return nil, HandleErrorResponse(res, err)
	}

	return workspace, nil
}

func GetFirstWorkspaceProjectName(workspaceId string, projectName string, profile *config.Profile) (string, error) {
	ctx := context.Background()

	apiClient, err := GetApiClient(profile)
	if err != nil {
		return "", err
	}

	wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceId).Execute()
	if err != nil {
		return "", HandleErrorResponse(res, err)
	}

	if projectName == "" {
		if len(wsInfo.Projects) == 0 {
			return "", errors.New("no projects found in workspace")
		}

		return *wsInfo.Projects[0].Name, nil
	}

	for _, project := range wsInfo.Projects {
		if *project.Name == projectName {
			return *project.Name, nil
		}
	}

	return "", errors.New("project not found in workspace")
}
