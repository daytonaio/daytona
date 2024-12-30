// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package frpc

import (
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/pkg/common"
	"github.com/fatedier/frp/client"
	v1 "github.com/fatedier/frp/pkg/config/v1"
)

type FrpcConnectParams struct {
	ServerDomain string
	ServerPort   int
	Name         string
	SubDomain    string
	Port         int
}

type HealthCheckFunc func() error

func GetService(params FrpcConnectParams) (HealthCheckFunc, *client.Service, error) {
	cfg := client.ServiceOptions{}
	cfg.Common = &v1.ClientCommonConfig{}
	cfg.Common.ServerAddr = params.ServerDomain
	cfg.Common.ServerPort = params.ServerPort
	cfg.ProxyCfgs = []v1.ProxyConfigurer{}

	httpConfig := &v1.HTTPProxyConfig{}
	httpConfig.GetBaseConfig().Name = params.Name
	httpConfig.GetBaseConfig().LocalPort = params.Port
	httpConfig.GetBaseConfig().Type = string(v1.ProxyTypeHTTP)
	httpConfig.SubDomain = params.SubDomain

	cfg.ProxyCfgs = append(cfg.ProxyCfgs, httpConfig)

	service, err := client.NewService(cfg)
	if err != nil {
		return nil, nil, err
	}

	return func() error {
		proxyStatus, ok := service.StatusExporter().GetProxyStatus(params.Name)
		if !ok || proxyStatus == nil {
			return fmt.Errorf("%w %w", errors.New("failed to get proxy status"), common.ErrConnection)
		}

		if proxyStatus.Err != "" {
			return fmt.Errorf("proxy error: %s %w", proxyStatus.Err, common.ErrConnection)
		}

		if proxyStatus.Phase != "running" {
			return fmt.Errorf("proxy state is not running. State is %s. %w", proxyStatus.Phase, common.ErrConnection)
		}

		return nil
	}, service, nil
}
