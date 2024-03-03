// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/frpc"
	"github.com/daytonaio/daytona/pkg/ports"
	serverfrpc "github.com/daytonaio/daytona/pkg/server/frpc"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var publicPreview bool

var portForwardCmd = &cobra.Command{
	Use:   "forward [WORKSPACE_NAME] [PROJECT_NAME] -p [PORT]",
	Short: "Forward port",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		workspaceId := args[0]
		projectName := ""
		var err error

		if len(args) == 2 {
			projectName = args[1]
		} else {
			projectName, err = util.GetFirstWorkspaceProjectName(workspaceId, projectName, nil)
			if err != nil {
				log.Fatal(err)
			}
		}

		hostPort, errChan := ports.ForwardPort(workspaceId, projectName, uint16(portArg))

		if hostPort == nil {
			if err = <-errChan; err != nil {
				log.Fatal(err)
			}
		} else {
			if *hostPort != uint16(portArg) {
				view_util.RenderInfoMessage(fmt.Sprintf("Port %d already in use.", portArg))
			}
			view_util.RenderInfoMessage(fmt.Sprintf("Port available at http://localhost:%d\n", *hostPort))
		}

		if publicPreview {
			go func() {
				errChan <- forwardPublicPort(workspaceId, projectName)
			}()
		}

		for {
			err := <-errChan
			if err != nil {
				log.Debug(err)
			}
		}
	},
}

func init() {
	portForwardCmd.Flags().BoolVar(&publicPreview, "public", false, "Should be port be available publicly via an URL")
	portForwardCmd.Flags().IntVarP(&portArg, "port", "p", 0, "Port to forward")
	portForwardCmd.MarkFlagRequired("port")
}

func forwardPublicPort(workspaceId, projectName string) error {
	log.Info("Forwarding port to a public URL...")

	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		return err
	}

	serverConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	// TODO: Remove uuid from the end when workspaceId becomes an uuid (currently, wsId is the same as its name)
	subDomain := fmt.Sprintf("%s-%s-%d-%s", workspaceId, projectName, portArg, uuid.New().String())

	go func() {
		time.Sleep(1 * time.Second)
		log.Infof("Port available at %s", fmt.Sprintf("%s://%s.%s", subDomain, serverfrpc.GetServerDomain(server.ToServerConfig(serverConfig)), *serverConfig.Frps.Protocol))
	}()

	return frpc.Connect(frpc.FrpcConnectParams{
		ServerDomain: *serverConfig.Frps.Domain,
		ServerPort:   int(*serverConfig.Frps.Port),
		Name:         subDomain,
		SubDomain:    subDomain,
	})
}
