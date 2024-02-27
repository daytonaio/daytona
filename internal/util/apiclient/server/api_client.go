package server

import (
	"net/http"
	"os"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/types"
)

var apiClient *serverapiclient.APIClient

func GetApiClient(profile *config.Profile) (*serverapiclient.APIClient, error) {
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

	clientConfig := serverapiclient.NewConfiguration()
	clientConfig.Servers = serverapiclient.ServerConfigurations{
		{
			URL: serverUrl,
		},
	}

	// clientConfig.AddDefaultHeader("Authorization", "Bearer "+token)

	apiClient = serverapiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	return apiClient, nil
}

func ToServerConfig(config *serverapiclient.ServerConfig) *types.ServerConfig {
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
