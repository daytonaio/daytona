// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"fmt"
	"strconv"

	defaultPortForwardCmd "github.com/daytonaio/daytona/pkg/cmd/ports"
	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/daytonaio/daytona/pkg/views"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var publicPreview bool

var portForwardCmd = &cobra.Command{
	Use:   "forward [PORT]",
	Short: "Forward a port from the project to your local machine",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		port, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}

		hostPort, errChan := ports.ForwardPort(workspaceId, projectName, uint16(port))

		if hostPort == nil {
			if err = <-errChan; err != nil {
				log.Fatal(err)
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
