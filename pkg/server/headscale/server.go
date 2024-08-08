// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package headscale

import (
	"fmt"
	"net"
	"os"

	"github.com/juanfont/headscale/hscontrol"
)

type HeadscaleServerConfig struct {
	ServerId      string
	FrpsDomain    string
	FrpsProtocol  string
	HeadscalePort uint32
	ConfigDir     string
}

func NewHeadscaleServer(config *HeadscaleServerConfig) *HeadscaleServer {
	return &HeadscaleServer{
		serverId:      config.ServerId,
		frpsDomain:    config.FrpsDomain,
		frpsProtocol:  config.FrpsProtocol,
		headscalePort: config.HeadscalePort,
		configDir:     config.ConfigDir,
	}
}

type HeadscaleServer struct {
	serverId      string
	frpsDomain    string
	frpsProtocol  string
	headscalePort uint32
	configDir     string
}

func (s *HeadscaleServer) Init() error {
	return os.MkdirAll(s.configDir, 0700)
}

func (s *HeadscaleServer) Start() error {
	_, err := net.Dial("tcp", fmt.Sprintf(":%d", s.headscalePort))
	if err == nil {
		return fmt.Errorf("cannot start Headscale server, port %d is already in use", s.headscalePort)
	}

	cfg, err := s.getHeadscaleConfig()
	if err != nil {
		return err
	}

	app, err := hscontrol.NewHeadscale(cfg)
	if err != nil {
		return err
	}

	return app.Serve()
}
