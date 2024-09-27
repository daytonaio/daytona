// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"strconv"

	"github.com/daytonaio/daytona/internal/util"
	defaultPortForwardCmd "github.com/daytonaio/daytona/pkg/cmd/ports"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var portForwardCmd = &cobra.Command{
	Use:     "forward [PORT]",
	Short:   "Forward a port publicly via an URL",
	Args:    cobra.ExactArgs(1),
	GroupID: util.WORKSPACE_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		errChan := make(chan error)
		go func() {
			errChan <- defaultPortForwardCmd.ForwardPublicPort(workspaceId, projectName, uint16(port), uint16(port))
		}()

		for {
			err := <-errChan
			if err != nil {
				log.Debug(err)
			}
		}
	},
}
