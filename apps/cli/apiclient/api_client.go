// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient

import (
	"context"
	"net/http"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/daytona/cli/auth"
	"github.com/daytonaio/daytona/cli/config"
)

var apiClient *apiclient.APIClient

const DaytonaSourceHeader = "X-Daytona-Source"

func GetApiClient(profile *config.Profile, defaultHeaders map[string]string) (*apiclient.APIClient, error) {
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

	if apiClient != nil && activeProfile.Api.Key == nil {
		err := auth.RefreshTokenIfNeeded(context.Background())
		if err != nil {
			return nil, err
		}

		return apiClient, nil
	}

	var newApiClient *apiclient.APIClient

	serverUrl := activeProfile.Api.Url

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: serverUrl,
		},
	}

	if activeProfile.Api.Key != nil {
		clientConfig.AddDefaultHeader("Authorization", "Bearer "+*activeProfile.Api.Key)
	} else if activeProfile.Api.Token != nil {
		clientConfig.AddDefaultHeader("Authorization", "Bearer "+activeProfile.Api.Token.AccessToken)

		if activeProfile.ActiveOrganizationId != nil {
			clientConfig.AddDefaultHeader("X-Daytona-Organization-ID", *activeProfile.ActiveOrganizationId)
		}
	}

	clientConfig.AddDefaultHeader(DaytonaSourceHeader, "cli")

	for headerKey, headerValue := range defaultHeaders {
		clientConfig.AddDefaultHeader(headerKey, headerValue)
	}

	newApiClient = apiclient.NewAPIClient(clientConfig)

	newApiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	if apiClient != nil && activeProfile.Api.Key == nil {
		err = auth.RefreshTokenIfNeeded(context.Background())
		if err != nil {
			return nil, err
		}
	}

	apiClient = newApiClient
	return apiClient, nil
}
