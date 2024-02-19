// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package connection

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cli/config"
	server_config "github.com/daytonaio/daytona/server/config"

	ssh_tunnel_util "github.com/daytonaio/daytona/pkg/ssh_tunnel/util"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// Get returns a grpc client connection to the local server or remote server
// based on the profile passed in. If no profile is passed in, the active profile
// is used.
func Get(profile *config.Profile) (*grpc.ClientConn, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	var activeProfile config.Profile
	if profile == nil {
		var err error
		activeProfile, err = c.GetActiveProfile()
		if err != nil {
			return nil, err
		}
	} else {
		activeProfile = *profile
	}

	if activeProfile.Id == "default" {
		localServerConfig, err := server_config.GetConfig()
		if err != nil {
			return nil, err
		}

		if localServerConfig == nil {
			return nil, errors.New("local server not configured. Run `daytona configure` first")
		}

		apiUrl := "127.0.0.1:3000"
		if serverApiUrl, ok := os.LookupEnv("DAYTONA_SERVER_API_URL"); ok {
			apiUrl = serverApiUrl
		}

		client, err := grpc.DialContext(ctx, apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		return client, err
	} else {
		sshTunnelContext, cancelTunnel := context.WithCancel(ctx)
		hostPort, errChan := ssh_tunnel_util.ForwardRemoteTcpPort(sshTunnelContext, activeProfile, 3000)

		go func() {
			if err := <-errChan; err != nil {
				log.Fatal(err)
			}
		}()

		client, err := grpc.DialContext(sshTunnelContext, fmt.Sprintf("localhost:%d", hostPort), grpc.WithTransportCredentials(insecure.NewCredentials()))

		go func() {
			for {
				if client.GetState() == connectivity.Shutdown {
					cancelTunnel()
					break
				}
			}
		}()

		return client, err
	}
}
