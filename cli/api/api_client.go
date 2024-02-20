package api

import (
	"net/http"

	"github.com/daytonaio/daytona/common/api_client"
)

var apiClient *api_client.APIClient

func GetServerApiClient(serverUrl, token string) *api_client.APIClient {
	if apiClient != nil {
		return apiClient
	}

	clientConfig := api_client.NewConfiguration()
	clientConfig.Servers = api_client.ServerConfigurations{
		{
			URL: serverUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+token)

	apiClient = api_client.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	return apiClient
}
