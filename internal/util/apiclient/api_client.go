// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/constants"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
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
		if common.AgentMode() {
			clientConfig.AddDefaultHeader(telemetry.SOURCE_HEADER, string(telemetry.CLI_WORKSPACE_SOURCE))
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

func GetRunnerApiClient(apiUrl, apiKey, clientId string, telemetryEnabled bool) (*apiclient.APIClient, error) {
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
		clientConfig.AddDefaultHeader(telemetry.SOURCE_HEADER, string(telemetry.RUNNER_SOURCE))
	}

	apiClient = apiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

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

func GetTarget(targetNameOrId string) (*apiclient.TargetDTO, int, error) {
	ctx := context.Background()

	apiClient, err := GetApiClient(nil)
	if err != nil {
		return nil, -1, err
	}

	target, res, err := apiClient.TargetAPI.GetTarget(ctx, targetNameOrId).Execute()
	if err != nil {
		return nil, res.StatusCode, HandleErrorResponse(res, err)
	}

	return target, res.StatusCode, nil
}

func GetWorkspace(workspaceNameOrId string) (*apiclient.WorkspaceDTO, int, error) {
	ctx := context.Background()

	apiClient, err := GetApiClient(nil)
	if err != nil {
		return nil, -1, err
	}

	workspace, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceNameOrId).Execute()
	if err != nil {
		return nil, res.StatusCode, HandleErrorResponse(res, err)
	}

	return workspace, res.StatusCode, nil
}
