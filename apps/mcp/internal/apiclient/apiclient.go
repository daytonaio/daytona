// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/daytonaio/apiclient"
	"github.com/daytonaio/toolbox_apiclient"
)

func NewApiClient(source, apiUrl string, header http.Header) *apiclient.APIClient {
	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: apiUrl,
		},
	}

	clientConfig.HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	clientConfig.AddDefaultHeader("Authorization", header.Get("Authorization"))
	clientConfig.AddDefaultHeader("X-Daytona-Source", source)

	return apiclient.NewAPIClient(clientConfig)
}

func ExtractProxyUrl(ctx context.Context, apiClient *apiclient.APIClient) (string, error) {
	config, _, err := apiClient.ConfigAPI.ConfigControllerGetConfig(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("error getting config: %v", err)
	}

	if config == nil {
		return "", errors.New("config is nil")
	}

	return config.ProxyToolboxUrl, nil
}

func NewToolboxApiClient(source, sandboxId, proxyUrl string, header http.Header) *toolbox_apiclient.APIClient {
	if !strings.HasSuffix(proxyUrl, "/") {
		proxyUrl = proxyUrl + "/"
	}
	proxyUrl = proxyUrl + sandboxId

	clientConfig := toolbox_apiclient.NewConfiguration()
	clientConfig.Servers = toolbox_apiclient.ServerConfigurations{
		{
			URL: proxyUrl,
		},
	}

	clientConfig.HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	clientConfig.AddDefaultHeader("Authorization", header.Get("Authorization"))
	clientConfig.AddDefaultHeader("X-Daytona-Source", source)

	return toolbox_apiclient.NewAPIClient(clientConfig)
}
