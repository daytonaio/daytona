// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daytona

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DaytonaMCPServer struct {
	*mcp.Server
}

func NewDaytonaMCPServer() *DaytonaMCPServer {
	server := &DaytonaMCPServer{
		Server: mcp.NewServer(&mcp.Implementation{
			Name:    "Daytona MCP Server",
			Version: internal.DaytonaMcpVersion,
		}, &mcp.ServerOptions{
			KeepAlive: 30 * time.Second,
			HasTools:  true,
		}),
	}

	server.addTools()

	return server
}

func (s *DaytonaMCPServer) Start(ctx context.Context, transport mcp.Transport) error {
	return s.Server.Run(ctx, transport)
}

func (s *DaytonaMCPServer) addTools() {
	mcp.AddTool(s.Server, getRunCodeTool(), handleRunCode)
	mcp.AddTool(s.Server, getShellTool(), handleShell)
}
