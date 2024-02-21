package api

import (
	"net/http"

	"github.com/daytonaio/daytona/common/api_client"
	"github.com/daytonaio/daytona/common/types"
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

func FromServerConfig(config *types.ServerConfig) *api_client.ServerConfig {
	return &api_client.ServerConfig{
		PluginsDir:        &config.PluginsDir,
		PluginRegistryUrl: &config.PluginRegistryUrl,
		Id:                &config.Id,
		ServerDownloadUrl: &config.ServerDownloadUrl,
		Frps: &api_client.FRPSConfig{
			Domain:   &config.Frps.Domain,
			Port:     &[]int32{int32(config.Frps.Port)}[0],
			Protocol: &config.Frps.Protocol,
		},
		ApiPort:       &[]int32{int32(config.ApiPort)}[0],
		HeadscalePort: &[]int32{int32(config.HeadscalePort)}[0],
	}
}
