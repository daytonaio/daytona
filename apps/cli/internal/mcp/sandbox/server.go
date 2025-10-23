// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"time"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewDaytonaSandboxMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "Daytona Sandbox MCP Server",
		Version: internal.SandboxMcpVersion,
	}, &mcp.ServerOptions{
		KeepAlive: 30 * time.Second,
		HasTools:  true,
	})

	addTools(server)

	return server
}

func addTools(server *mcp.Server) {
	mcp.AddTool(server, getCreateSandboxTool(), handleCreateSandbox)
	mcp.AddTool(server, getDestroySandboxTool(), handleDestroySandbox)
	mcp.AddTool(server, getPreviewLinkTool(), handlePreviewLink)
}
