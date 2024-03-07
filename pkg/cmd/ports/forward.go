// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports

import (
	"context"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"time"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/frpc"
	"github.com/daytonaio/daytona/pkg/ports"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var publicPreview bool
var workspaceId string
var projectName string

var PortForwardCmd = &cobra.Command{
	Use:   "forward [PORT]",
	Short: "Forward a port from a project to your local machine",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		port, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}

		if len(args) > 1 {
			workspace, err := server.GetWorkspace(args[1])
			if err != nil {
				log.Fatal(err)
			}
			workspaceId = *workspace.Id
		}

		if len(args) == 3 {
			projectName = args[2]
		} else {
			projectName, err = util.GetFirstWorkspaceProjectName(workspaceId, projectName, nil)
			if err != nil {
				log.Fatal(err)
			}
		}

		hostPort, errChan := ports.ForwardPort(workspaceId, projectName, uint16(port))

		if hostPort == nil {
			if err = <-errChan; err != nil {
				log.Fatal(err)
			}
		} else {
			if *hostPort != uint16(port) {
				view_util.RenderInfoMessage(fmt.Sprintf("Port %d already in use.", port))
			}
			view_util.RenderInfoMessage(fmt.Sprintf("Port available at http://localhost:%d\n", *hostPort))
		}

		if publicPreview {
			go func() {
				errChan <- forwardPublicPort(workspaceId, projectName, *hostPort, uint16(port))
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
	if wsId := os.Getenv("DAYTONA_WS_ID"); wsId != "" {
		workspaceId = wsId
	}
	if pName := os.Getenv("DAYTONA_PROJECT_NAME"); pName != "" {
		projectName = pName
	}

	if !util.WorkspaceMode() {
		PortForwardCmd.Use = PortForwardCmd.Use + " [WORKSPACE] [PROJECT]"
		PortForwardCmd.Args = cobra.RangeArgs(2, 3)
	}

	PortForwardCmd.Flags().BoolVar(&publicPreview, "public", false, "Should be port be available publicly via an URL")
}

func forwardPublicPort(workspaceId, projectName string, hostPort, targetPort uint16) error {
	view_util.RenderInfoMessage("Forwarding port to a public URL...")

	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		return err
	}

	serverConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	h := fnv.New64()
	h.Write([]byte(fmt.Sprintf("%s-%s-%s", workspaceId, projectName, *serverConfig.Id)))

	subDomain := fmt.Sprintf("%d-%s", targetPort, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprint(h.Sum64()))))

	go func() {
		time.Sleep(1 * time.Second)
		view_util.RenderInfoMessage(fmt.Sprintf("Port available at %s", fmt.Sprintf("%s://%s.%s", *serverConfig.Frps.Protocol, subDomain, *serverConfig.Frps.Domain)))
	}()

	return frpc.Connect(frpc.FrpcConnectParams{
		ServerDomain: *serverConfig.Frps.Domain,
		ServerPort:   int(*serverConfig.Frps.Port),
		Name:         subDomain,
		SubDomain:    subDomain,
		Port:         int(hostPort),
	})
}
