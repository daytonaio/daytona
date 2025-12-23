// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient

import (
	"net/http"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner/cmd/runner/config"
)

var apiClient *apiclient.APIClient

const DaytonaSourceHeader = "X-Daytona-Source"

func GetApiClient() (*apiclient.APIClient, error) {

	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	var newApiClient *apiclient.APIClient

	serverUrl := c.DaytonaApiUrl

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: serverUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+c.ApiToken)

	clientConfig.AddDefaultHeader(DaytonaSourceHeader, "runner")

	newApiClient = apiclient.NewAPIClient(clientConfig)

	newApiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	apiClient = newApiClient
	return apiClient, nil
}
