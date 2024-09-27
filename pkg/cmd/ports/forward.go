// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/cmd/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/frpc"
	"github.com/daytonaio/daytona/pkg/views"
	log "github.com/sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var publicPreview bool
var workspaceId string
var projectName string

var PortForwardCmd = &cobra.Command{
	Use:     "forward [PORT] [WORKSPACE] [PROJECT]",
	Short:   "Forward a port from a project to your local machine",
	GroupID: util.WORKSPACE_GROUP,
	Args:    cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}
		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		port, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		workspace, err := apiclient.GetWorkspace(args[1], true)
		if err != nil {
			return err
		}
		workspaceId = workspace.Id

		if len(args) == 3 {
			projectName = args[2]
		} else {
			projectName, err = apiclient.GetFirstWorkspaceProjectName(workspaceId, projectName, nil)
			if err != nil {
				return err
			}
		}

		hostPort, errChan := tailscale.ForwardPort(workspaceId, projectName, uint16(port), activeProfile)

		if hostPort == nil {
			if err = <-errChan; err != nil {
				return err
			}
		} else {
			if *hostPort != uint16(port) {
				views.RenderInfoMessage(fmt.Sprintf("Port %d already in use.", port))
			}
			views.RenderInfoMessage(fmt.Sprintf("Port available at http://localhost:%d\n", *hostPort))
		}

		if publicPreview {
			go func() {
				errChan <- ForwardPublicPort(workspaceId, projectName, *hostPort, uint16(port))
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
	PortForwardCmd.Flags().BoolVar(&publicPreview, "public", false, "Should be port be available publicly via an URL")
}

func ForwardPublicPort(workspaceId, projectName string, hostPort, targetPort uint16) error {
	views.RenderInfoMessage("Forwarding port to a public URL...")

	apiClient, err := apiclient.GetApiClient(nil)
	if err != nil {
		return err
	}

	serverConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	h := fnv.New64()
	h.Write([]byte(fmt.Sprintf("%s-%s-%s", workspaceId, projectName, serverConfig.Id)))

	subDomain := fmt.Sprintf("%d-%s", targetPort, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprint(h.Sum64()))))

	if serverConfig.Frps == nil {
		return errors.New("frps config is missing")
	}

	go func() {
		time.Sleep(1 * time.Second)
		var url = fmt.Sprintf("%s://%s.%s", serverConfig.Frps.Protocol, subDomain, serverConfig.Frps.Domain)
		views.RenderInfoMessage(fmt.Sprintf("Port available at %s", url))
		err := renderQr(url)
		if err != nil {
			log.Error(err)
		}
	}()

	_, service, err := frpc.GetService(frpc.FrpcConnectParams{
		ServerDomain: serverConfig.Frps.Domain,
		ServerPort:   int(serverConfig.Frps.Port),
		Name:         subDomain,
		SubDomain:    subDomain,
		Port:         int(hostPort),
	})
	if err != nil {
		return err
	}

	return service.Run(context.Background())
}

func renderQr(s string) error {
	q, err := qrcode.New(s, qrcode.Medium)
	if err != nil {
		return err
	}
	fmt.Println(q.ToSmallString(true))
	return nil
}
