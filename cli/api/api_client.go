package api

import (
	"net/http"
	"os"

	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/common/api_client"
	"github.com/daytonaio/daytona/common/types"
)

var apiClient *api_client.APIClient

func GetServerApiClient(profile *config.Profile) (*api_client.APIClient, error) {
	if apiClient != nil {
		return apiClient, nil
	}

	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	serverUrl := "http://localhost:3000"

	if envApiUrl, ok := os.LookupEnv("DAYTONA_SERVER_API_URL"); ok {
		serverUrl = envApiUrl
	} else {
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

		serverUrl = activeProfile.Api.Url
	}

	clientConfig := api_client.NewConfiguration()
	clientConfig.Servers = api_client.ServerConfigurations{
		{
			URL: serverUrl,
		},
	}

	// clientConfig.AddDefaultHeader("Authorization", "Bearer "+token)

	apiClient = api_client.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	return apiClient, nil
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
