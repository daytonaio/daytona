// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"time"

	"github.com/daytonaio/apiclient"

	"github.com/daytonaio/mcp/internal"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DaytonaFileSystemMCPServer struct {
	*mcp.Server
	apiClient *apiclient.APIClient
}

func NewDaytonaFileSystemMCPServer(apiClient *apiclient.APIClient) *DaytonaFileSystemMCPServer {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "Daytona FileSystem MCP Server",
		Version: internal.FsMcpVersion,
	}, &mcp.ServerOptions{
		KeepAlive: 30 * time.Second,
		HasTools:  true,
	})

	server := &DaytonaFileSystemMCPServer{
		Server:    mcpServer,
		apiClient: apiClient,
	}

	server.addTools()

	return server
}

func (s *DaytonaFileSystemMCPServer) addTools() {
	mcp.AddTool(s.Server, s.getCreateFolderTool(), s.handleCreateFolder)
	mcp.AddTool(s.Server, s.getDeleteFileTool(), s.handleDeleteFile)
	mcp.AddTool(s.Server, s.getDownloadFileTool(), s.handleDownloadFile)
	mcp.AddTool(s.Server, s.getFileInfoTool(), s.handleFileInfo)
	mcp.AddTool(s.Server, s.getListFilesTool(), s.handleListFiles)
	mcp.AddTool(s.Server, s.getMoveFileTool(), s.handleMoveFile)
	mcp.AddTool(s.Server, s.getUploadFileTool(), s.handleUploadFile)
}
