// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mcp

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/daytonaio/daytona/cli/apiclient"

	"github.com/daytonaio/mcp/pkg/servers/daytona"
	"github.com/daytonaio/mcp/pkg/servers/fs"
	"github.com/daytonaio/mcp/pkg/servers/git"
	"github.com/daytonaio/mcp/pkg/servers/sandbox"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve any Daytona MCP Server over STDIO (empty string for daytona code execution MCP, 'sandbox' for Sandbox actions MCP, 'fs' for Filesystem operations MCP, 'git' for Git operations MCP)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mcpServerName := ""
		if len(args) == 1 {
			mcpServerName = args[0]
		}

		apiClient, err := apiclient.GetApiClient(nil, map[string]string{apiclient.DaytonaSourceHeader: "daytona-mcp-stdio"})
		if err != nil {
			return err
		}

		var server *mcp.Server
		switch mcpServerName {
		case "":
			server = daytona.NewDaytonaMCPServer(apiClient).Server
		case "sandbox":
			server = sandbox.NewDaytonaSandboxMCPServer(apiClient).Server
		case "fs":
			server = fs.NewDaytonaFileSystemMCPServer(apiClient).Server
		case "git":
			server = git.NewDaytonaGitMCPServer(apiClient).Server
		default:
			return fmt.Errorf("mcp server name %s is not supported", mcpServerName)
		}

		interruptChan := make(chan os.Signal, 1)
		signal.Notify(interruptChan, os.Interrupt)

		errChan := make(chan error)

		go func() {
			errChan <- server.Run(context.Background(), &mcp.StdioTransport{})
		}()

		select {
		case err := <-errChan:
			return err
		case <-interruptChan:
			return nil
		}
	},
}
