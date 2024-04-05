// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package headscale

import (
	"fmt"
	"net/http"

	"tailscale.com/tsnet"

	"github.com/daytonaio/daytona/internal/util"
	log "github.com/sirupsen/logrus"
)

var tsNetServer = &tsnet.Server{
	Hostname: "server",
}

func (s *HeadscaleServer) Connect() error {
	err := s.CreateUser()
	if err != nil {
		log.Fatal(err)
	}

	authKey, err := s.CreateAuthKey()
	if err != nil {
		log.Fatal(err)
	}

	tsNetServer.ControlURL = util.GetFrpcServerUrl(s.frpsProtocol, s.serverId, s.frpsDomain)
	tsNetServer.AuthKey = authKey

	defer tsNetServer.Close()
	ln, err := tsNetServer.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	return http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Ok\n")
	}))
}

func (s *HeadscaleServer) HTTPClient() *http.Client {
	return tsNetServer.HTTPClient()
}
