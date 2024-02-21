// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package connection

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/api"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/common/api_client"
	"github.com/daytonaio/daytona/server/frpc"
	"github.com/google/uuid"
	"tailscale.com/tsnet"

	log "github.com/sirupsen/logrus"
)

var s *tsnet.Server = nil

func GetTailscaleConn(profile *config.Profile) (*tsnet.Server, error) {
	if s != nil {
		return s, nil
	}
	s = &tsnet.Server{}

	apiClient := api.GetServerApiClient("http://localhost:3000", "")

	serverConfig, _, err := apiClient.ServerAPI.GetConfigExecute(api_client.ApiGetConfigRequest{})
	if err != nil {
		log.Fatal(err)
	}

	networkKey, _, err := apiClient.ServerAPI.GenerateNetworkKeyExecute(api_client.ApiGenerateNetworkKeyRequest{})
	if err != nil {
		return nil, err
	}

	s.Hostname = fmt.Sprintf("cli-%s", uuid.New().String())
	s.ControlURL = frpc.GetServerUrl(api.ToServerConfig(serverConfig))
	s.AuthKey = *networkKey.Key
	s.Ephemeral = true
	s.Logf = func(format string, args ...any) {}

	_, err = s.Up(context.Background())
	if err != nil {
		return nil, err
	}

	return s, nil
}
