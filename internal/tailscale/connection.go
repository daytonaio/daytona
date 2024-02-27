// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package tailscale

import (
	"context"
	"fmt"
	"path"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/server/frpc"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/google/uuid"
	"tailscale.com/tsnet"
)

var s *tsnet.Server = nil

func GetConnection(profile *config.Profile) (*tsnet.Server, error) {
	if s != nil {
		return s, nil
	}
	s = &tsnet.Server{}

	apiClient, err := server.GetApiClient(profile)
	if err != nil {
		return nil, err
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(serverapiclient.ApiGetConfigRequest{})
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	networkKey, res, err := apiClient.ServerAPI.GenerateNetworkKeyExecute(serverapiclient.ApiGenerateNetworkKeyRequest{})
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	cliId := uuid.New().String()
	s.Hostname = fmt.Sprintf("cli-%s", cliId)
	s.ControlURL = frpc.GetServerUrl(server.ToServerConfig(serverConfig))
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
