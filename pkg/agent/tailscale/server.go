// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package tailscale

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"tailscale.com/tsnet"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	Hostname string
	Server   config.DaytonaServerConfig
}

func (s *Server) Start() error {
	flag.Parse()
	tsnetServer := new(tsnet.Server)
	tsnetServer.Hostname = s.Hostname
	tsnetServer.ControlURL = s.Server.Url
	tsnetServer.Ephemeral = true

	apiClient, err := server.GetAgentApiClient(s.Server.ApiUrl, s.Server.ApiKey)
	if err != nil {
		return err
	}

	networkKey, res, err := apiClient.ServerAPI.GenerateNetworkKeyExecute(serverapiclient.ApiGenerateNetworkKeyRequest{})
	if err != nil {
		log.Fatal(apiclient.HandleErrorResponse(res, err))
	}

	tsnetServer.AuthKey = *networkKey.Key

	defer tsnetServer.Close()
	ln, err := tsnetServer.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	tsnetServer.RegisterFallbackTCPHandler(func(src, dest netip.AddrPort) (handler func(net.Conn), intercept bool) {
		destPort := dest.Port()

		return func(src net.Conn) {
			defer src.Close()
			dst, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", destPort))
			if err != nil {
				log.Errorf("Dial failed: %v", err)
				return
			}
			defer dst.Close()

			done := make(chan struct{})

			go func() {
				defer src.Close()
				defer dst.Close()
				io.Copy(dst, src)
				done <- struct{}{}
			}()

			go func() {
				defer src.Close()
				defer dst.Close()
				io.Copy(src, dst)
				done <- struct{}{}
			}()

			<-done
			<-done
		}, true
	})

	return http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Ok\n")
	}))
}
