package frpc

import (
	"context"

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

func Connect(params FrpcConnectParams) error {
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
		return err
	}

	return service.Run(context.Background())
}
