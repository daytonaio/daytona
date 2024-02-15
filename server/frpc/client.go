package frpc

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/common/grpc/proto/types"
	"github.com/daytonaio/daytona/server/config"
	"github.com/fatedier/frp/client"
	v1 "github.com/fatedier/frp/pkg/config/v1"
)

func newService() (*client.Service, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	serverDomain := GetServerDomain(c)

	cfg := client.ServiceOptions{}
	cfg.Common = &v1.ClientCommonConfig{}
	cfg.Common.ServerAddr = c.Frps.Domain
	cfg.Common.ServerPort = int(c.Frps.Port)
	cfg.ProxyCfgs = []v1.ProxyConfigurer{}

	httpConfig := &v1.HTTPProxyConfig{}
	httpConfig.GetBaseConfig().Name = "daytona-server"
	httpConfig.GetBaseConfig().LocalPort = int(c.HeadscalePort)
	httpConfig.GetBaseConfig().Type = string(v1.ProxyTypeHTTP)
	httpConfig.CustomDomains = []string{serverDomain}

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

func GetServerDomain(c *types.ServerConfig) string {
	return fmt.Sprintf("%s.%s", c.Id, c.Frps.Domain)
}

func GetServerUrl(c *types.ServerConfig) string {
	return fmt.Sprintf("%s://%s.%s", c.Frps.Protocol, c.Id, c.Frps.Domain)
}
