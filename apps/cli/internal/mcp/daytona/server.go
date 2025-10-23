// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daytona

import (
	"time"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewDaytonaMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "Daytona MCP Server",
		Version: internal.DaytonaMcpVersion,
	}, &mcp.ServerOptions{
		KeepAlive: 30 * time.Second,
		HasTools:  true,
	})

	addTools(server)

	return server
}

func addTools(server *mcp.Server) {
	mcp.AddTool(server, getRunCodeTool(), handleRunCode)
	mcp.AddTool(server, getShellTool(), handleShell)
}
