// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package tailscale

import (
	"fmt"
	"path/filepath"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/tailscale"
	"github.com/google/uuid"
	"tailscale.com/tsnet"
)

func GetConnection(profile *config.Profile) (*tsnet.Server, error) {
	apiClient, err := server.GetApiClient(profile)
	if err != nil {
		return nil, err
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	networkKey, res, err := apiClient.ServerAPI.GenerateNetworkKeyExecute(apiclient.ApiGenerateNetworkKeyRequest{})
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	cliId := uuid.New().String()

	return tailscale.GetConnection(&tailscale.TsnetConnConfig{
		AuthKey:    *networkKey.Key,
		ControlURL: util.GetFrpcServerUrl(*serverConfig.Frps.Protocol, *serverConfig.Id, *serverConfig.Frps.Domain),
		Dir:        filepath.Join(configDir, "tailscale", cliId),
		Logf:       func(format string, args ...any) {},
		Hostname:   fmt.Sprintf("cli-%s", cliId),
	})
}
