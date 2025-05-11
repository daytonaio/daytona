// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mcp

import (
	"os"
	"os/signal"

	"github.com/daytonaio/daytona/cli/mcp"
	"github.com/spf13/cobra"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Daytona MCP Server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		server := mcp.NewDaytonaMCPServer()

		interruptChan := make(chan os.Signal, 1)
		signal.Notify(interruptChan, os.Interrupt)

		errChan := make(chan error)

		go func() {
			errChan <- server.Start()
		}()

		select {
		case err := <-errChan:
			return err
		case <-interruptChan:
			return nil
		}
	},
}
