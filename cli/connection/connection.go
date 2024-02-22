// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package connection

import (
	"context"
	"fmt"
	"path"

	"github.com/daytonaio/daytona/cli/api"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/common/api_client"
	"github.com/daytonaio/daytona/server/frpc"
	"github.com/google/uuid"
	"tailscale.com/tsnet"
)

var s *tsnet.Server = nil

func GetTailscaleConn(profile *config.Profile) (*tsnet.Server, error) {
	if s != nil {
		return s, nil
	}
	s = &tsnet.Server{}

	apiClient, err := api.GetServerApiClient(profile)
	if err != nil {
		return nil, err
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	serverConfig, _, err := apiClient.ServerAPI.GetConfigExecute(api_client.ApiGetConfigRequest{})
	if err != nil {
		return nil, err
	}

	networkKey, _, err := apiClient.ServerAPI.GenerateNetworkKeyExecute(api_client.ApiGenerateNetworkKeyRequest{})
	if err != nil {
		return nil, err
	}

	cliId := uuid.New().String()
	s.Hostname = fmt.Sprintf("cli-%s", cliId)
	s.ControlURL = frpc.GetServerUrl(api.ToServerConfig(serverConfig))
	s.AuthKey = *networkKey.Key
	s.Ephemeral = true
	s.Dir = path.Join(configDir, "tailscale", cliId)
	s.Logf = func(format string, args ...any) {}

	_, err = s.Up(context.Background())
	if err != nil {
		return nil, err
	}

	return s, nil
}
