// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package frpc

import (
	"fmt"

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
		proxyStatus, err := service.GetProxyStatus(params.Name)
		if err != nil {
			return err
		}

		if proxyStatus.Err != "" {
			return fmt.Errorf("proxy error: %s", proxyStatus.Err)
		}

		if proxyStatus.Phase != "running" {
			return fmt.Errorf("proxy state is not running. State is %s", proxyStatus.Phase)
		}

		return nil
	}, service, nil
}
