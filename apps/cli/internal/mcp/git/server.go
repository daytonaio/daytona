// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"time"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewDaytonaGitMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "Daytona Git MCP Server",
		Version: internal.GitMcpVersion,
	}, &mcp.ServerOptions{
		KeepAlive: 30 * time.Second,
		HasTools:  true,
	})

	addTools(server)

	return server
}

func addTools(server *mcp.Server) {
	mcp.AddTool(server, getGitCloneTool(), handleGitClone)
}
