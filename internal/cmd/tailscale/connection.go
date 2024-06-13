// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package tailscale

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/tailscale"
	"github.com/google/uuid"
	"tailscale.com/tsnet"

	log "github.com/sirupsen/logrus"
)

func GetConnection(profile *config.Profile) (*tsnet.Server, error) {
	apiClient, err := apiclient_util.GetApiClient(profile)
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

	var controlURL string
	if strings.Contains(profile.Api.Url, "localhost") || strings.Contains(profile.Api.Url, "0.0.0.0") || strings.Contains(profile.Api.Url, "127.0.0.1") {
		controlURL = fmt.Sprintf("http://localhost:%d", serverConfig.HeadscalePort)
	} else {
		if serverConfig.Frps == nil {
			return nil, fmt.Errorf("frps config is missing")
		}
		controlURL = util.GetFrpcHeadscaleUrl(serverConfig.Frps.Protocol, serverConfig.Id, serverConfig.Frps.Domain)
	}

	return tailscale.GetConnection(&tailscale.TsnetConnConfig{
		AuthKey:    networkKey.Key,
		ControlURL: controlURL,
		Dir:        filepath.Join(configDir, "tailscale", cliId),
		Logf: func(format string, args ...any) {
			log.Tracef(format, args...)
		},
		Hostname: fmt.Sprintf("cli-%s", cliId),
	})
}
