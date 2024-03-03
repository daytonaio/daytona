// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package headscale

import (
	"fmt"
	"net/http"

	"tailscale.com/tsnet"

	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/frpc"
	log "github.com/sirupsen/logrus"
)

var s = &tsnet.Server{
	Hostname: "server",
}

func Connect() error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	err = CreateUser()
	if err != nil {
		log.Fatal(err)
	}

	authKey, err := CreateAuthKey()
	if err != nil {
		log.Fatal(err)
	}

	s.ControlURL = frpc.GetServerUrl(c)
	s.AuthKey = authKey

	defer s.Close()
	ln, err := s.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	return http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Ok\n")
	}))
}

func HTTPClient() *http.Client {
	return s.HTTPClient()
}
