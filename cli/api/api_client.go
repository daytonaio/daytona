package api

import (
	"net/http"
	"os"

	"github.com/daytonaio/daytona/common/api_client"
	"github.com/daytonaio/daytona/common/types"
)

var apiClient *api_client.APIClient

func GetServerApiClient(serverUrl, token string) *api_client.APIClient {
	if apiClient != nil {
		return apiClient
	}

	if envApiUrl, ok := os.LookupEnv("DAYTONA_SERVER_API_URL"); ok {
		serverUrl = envApiUrl
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

func ToServerConfig(config *api_client.ServerConfig) *types.ServerConfig {
	return &types.ServerConfig{
		PluginsDir:        *config.PluginsDir,
		PluginRegistryUrl: *config.PluginRegistryUrl,
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
