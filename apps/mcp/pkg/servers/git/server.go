// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"time"

	apiclient "github.com/daytonaio/apiclient"

	"github.com/daytonaio/mcp/internal"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DaytonaGitMCPServer struct {
	*mcp.Server
	apiClient *apiclient.APIClient
}

func NewDaytonaGitMCPServer(apiClient *apiclient.APIClient) *DaytonaGitMCPServer {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "Daytona Git MCP Server",
		Version: internal.GitMcpVersion,
	}, &mcp.ServerOptions{
		KeepAlive: 30 * time.Second,
		HasTools:  true,
	})

	server := &DaytonaGitMCPServer{
		Server:    mcpServer,
		apiClient: apiClient,
	}

	server.addTools()

	return server
}

func (s *DaytonaGitMCPServer) addTools() {
	mcp.AddTool(s.Server, s.getGitCloneTool(), s.handleGitClone)
}
