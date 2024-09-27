// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"fmt"
	"strconv"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/cmd/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	defaultPortForwardCmd "github.com/daytonaio/daytona/pkg/cmd/ports"
	"github.com/daytonaio/daytona/pkg/views"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var publicPreview bool

var portForwardCmd = &cobra.Command{
	Use:     "forward [PORT]",
	Short:   "Forward a port from the project to your local machine",
	Args:    cobra.ExactArgs(1),
	GroupID: util.WORKSPACE_GROUP,
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
				errChan <- defaultPortForwardCmd.ForwardPublicPort(workspaceId, projectName, *hostPort, uint16(port))
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
}
