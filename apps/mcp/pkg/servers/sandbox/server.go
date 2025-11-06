// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"time"

	"github.com/daytonaio/apiclient"
	"github.com/daytonaio/mcp/internal"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DaytonaSandboxMCPServer struct {
	*mcp.Server
	apiClient *apiclient.APIClient
}

func NewDaytonaSandboxMCPServer(apiClient *apiclient.APIClient) *DaytonaSandboxMCPServer {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "Daytona Sandbox MCP Server",
		Version: internal.SandboxMcpVersion,
	}, &mcp.ServerOptions{
		KeepAlive: 30 * time.Second,
		HasTools:  true,
	})

	server := &DaytonaSandboxMCPServer{
		Server:    mcpServer,
		apiClient: apiClient,
	}

	server.addTools()

	return server
}

func (s *DaytonaSandboxMCPServer) addTools() {
	mcp.AddTool(s.Server, s.getCreateSandboxTool(), s.handleCreateSandbox)
	mcp.AddTool(s.Server, s.getDestroySandboxTool(), s.handleDestroySandbox)
	mcp.AddTool(s.Server, s.getPreviewLinkTool(), s.handlePreviewLink)
}
