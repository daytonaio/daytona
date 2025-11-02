// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient

import (
	"net/http"

	"github.com/daytonaio/apiclient"
)

func NewApiClient(apiUrl string, headers map[string]string) *apiclient.APIClient {
	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: apiUrl,
		},
	}

	clientConfig.HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	for headerKey, headerValue := range headers {
		clientConfig.AddDefaultHeader(headerKey, headerValue)
	}

	return apiclient.NewAPIClient(clientConfig)
}
