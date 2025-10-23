// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DaytonaGitMCPServer struct {
	*mcp.Server
}

func NewDaytonaGitMCPServer() *DaytonaGitMCPServer {
	server := &DaytonaGitMCPServer{
		Server: mcp.NewServer(&mcp.Implementation{
			Name:    "Daytona Git MCP Server",
			Version: internal.GitMcpVersion,
		}, &mcp.ServerOptions{
			KeepAlive: 30 * time.Second,
			HasTools:  true,
		}),
	}

	server.addTools()

	return server
}

func (s *DaytonaGitMCPServer) Start(ctx context.Context, transport mcp.Transport) error {
	return s.Server.Run(ctx, transport)
}

func (s *DaytonaGitMCPServer) addTools() {
	mcp.AddTool(s.Server, getGitCloneTool(), handleGitClone)
}
