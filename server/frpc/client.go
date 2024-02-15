package frpc

import (
	"context"

	"github.com/fatedier/frp/client"
	v1 "github.com/fatedier/frp/pkg/config/v1"
)

func newService() (*client.Service, error) {
	cfg := client.ServiceOptions{}
	cfg.Common = &v1.ClientCommonConfig{}
	cfg.Common.ServerAddr = "frps.daytona.io"
	cfg.Common.ServerPort = 7000
	cfg.ProxyCfgs = []v1.ProxyConfigurer{}

	httpConfig := &v1.HTTPProxyConfig{}
	httpConfig.GetBaseConfig().Name = "toma"
	httpConfig.GetBaseConfig().LocalPort = 8000
	httpConfig.GetBaseConfig().Type = string(v1.ProxyTypeHTTP)
	httpConfig.CustomDomains = []string{"toma.frps.daytona.io"}

	cfg.ProxyCfgs = append(cfg.ProxyCfgs, httpConfig)

	return client.NewService(cfg)
}

func Connect() error {
	service, err := newService()
	if err != nil {
		return err
	}

	return service.Run(context.Background())
}
