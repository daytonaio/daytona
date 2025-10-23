// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DaytonaSandboxMCPServer struct {
	*mcp.Server
}

func NewDaytonaSandboxMCPServer() *DaytonaSandboxMCPServer {
	server := &DaytonaSandboxMCPServer{
		Server: mcp.NewServer(&mcp.Implementation{
			Name:    "Daytona Sandbox MCP Server",
			Version: internal.SandboxMcpVersion,
		}, &mcp.ServerOptions{
			KeepAlive: 30 * time.Second,
			HasTools:  true,
		}),
	}

	server.addTools()

	return server
}

func (s *DaytonaSandboxMCPServer) Start(ctx context.Context, transport mcp.Transport) error {
	return s.Server.Run(ctx, transport)
}

func (s *DaytonaSandboxMCPServer) addTools() {
	mcp.AddTool(s.Server, getCreateSandboxTool(), handleCreateSandbox)
	mcp.AddTool(s.Server, getDestroySandboxTool(), handleDestroySandbox)
	mcp.AddTool(s.Server, getPreviewLinkTool(), handlePreviewLink)
}
