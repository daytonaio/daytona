// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daytona

import (
	"time"

	"github.com/daytonaio/apiclient"
	"github.com/daytonaio/mcp/internal"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DaytonaMCPServer struct {
	*mcp.Server
	apiClient *apiclient.APIClient
}

func NewDaytonaMCPServer(apiClient *apiclient.APIClient) *DaytonaMCPServer {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "Daytona MCP Server",
		Version: internal.DaytonaMcpVersion,
	}, &mcp.ServerOptions{
		KeepAlive: 30 * time.Second,
		HasTools:  true,
	})

	server := &DaytonaMCPServer{
		Server:    mcpServer,
		apiClient: apiClient,
	}

	server.addTools()

	return server
}

func (s *DaytonaMCPServer) addTools() {
	mcp.AddTool(s.Server, s.getRunCodeTool(), s.handleRunCode)
	mcp.AddTool(s.Server, s.getShellTool(), s.handleShell)
}
