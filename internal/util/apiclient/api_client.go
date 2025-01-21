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
	"github.com/daytonaio/daytona/pkg/telemetry"
)

const CLIENT_VERSION_HEADER = "X-Client-Version"

var apiClient *apiclient.APIClient

func GetApiClient(profile *config.Profile) (*apiclient.APIClient, error) {
	if apiClient != nil {
		return apiClient, nil
	}

	var newApiClient *apiclient.APIClient

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

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: serverUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	clientConfig.AddDefaultHeader(CLIENT_VERSION_HEADER, internal.Version)

	if c.TelemetryEnabled {
		clientConfig.AddDefaultHeader(telemetry.ENABLED_HEADER, "true")
		clientConfig.AddDefaultHeader(telemetry.SESSION_ID_HEADER, internal.SESSION_ID)
		clientConfig.AddDefaultHeader(telemetry.CLIENT_ID_HEADER, config.GetClientId())
		if internal.AgentMode() {
			clientConfig.AddDefaultHeader(telemetry.SOURCE_HEADER, string(telemetry.CLI_PROJECT_SOURCE))
		} else {
			clientConfig.AddDefaultHeader(telemetry.SOURCE_HEADER, string(telemetry.CLI_SOURCE))
		}
	}

	newApiClient = apiclient.NewAPIClient(clientConfig)

	newApiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	healthUrl, err := url.JoinPath(serverUrl, constants.HEALTH_CHECK_ROUTE)
	if err != nil {
		return nil, err
	}

	_, _, err = newApiClient.DefaultAPI.HealthCheck(context.Background()).Execute()
	if err != nil {
		return nil, ErrHealthCheckFailed(healthUrl)
	}

	apiClient = newApiClient
	return apiClient, nil
}

func GetAgentApiClient(apiUrl, apiKey, clientId string, telemetryEnabled bool) (*apiclient.APIClient, error) {
	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: apiUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	clientConfig.AddDefaultHeader(CLIENT_VERSION_HEADER, internal.Version)

	if telemetryEnabled {
		clientConfig.AddDefaultHeader(telemetry.ENABLED_HEADER, "true")
		clientConfig.AddDefaultHeader(telemetry.SESSION_ID_HEADER, internal.SESSION_ID)
		clientConfig.AddDefaultHeader(telemetry.CLIENT_ID_HEADER, clientId)
		clientConfig.AddDefaultHeader(telemetry.SOURCE_HEADER, string(telemetry.AGENT_SOURCE))
	}

	apiClient = apiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	return apiClient, nil
}

func GetTarget(targetNameOrId string, verbose bool) (*apiclient.TargetDTO, error) {
	ctx := context.Background()

	apiClient, err := GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	target, res, err := apiClient.TargetAPI.GetTarget(ctx, targetNameOrId).Verbose(verbose).Execute()
	if err != nil {
		return nil, HandleErrorResponse(res, err)
	}

	return target, nil
}

func GetFirstProjectName(targetId string, projectName string, profile *config.Profile) (string, error) {
	ctx := context.Background()

	apiClient, err := GetApiClient(profile)
	if err != nil {
		return "", err
	}

	targetInfo, res, err := apiClient.TargetAPI.GetTarget(ctx, targetId).Execute()
	if err != nil {
		return "", HandleErrorResponse(res, err)
	}

	if projectName == "" {
		if len(targetInfo.Projects) == 0 {
			return "", errors.New("no projects found in target")
		}

		return targetInfo.Projects[0].Name, nil
	}

	for _, project := range targetInfo.Projects {
		if project.Name == projectName {
			return project.Name, nil
		}
	}

	return "", errors.New("project not found in target")
}
