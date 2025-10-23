// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DaytonaFileSystemMCPServer struct {
	*mcp.Server
}

func NewDaytonaFileSystemMCPServer() *DaytonaFileSystemMCPServer {
	server := &DaytonaFileSystemMCPServer{
		Server: mcp.NewServer(&mcp.Implementation{
			Name:    "Daytona FileSystem MCP Server",
			Version: internal.FsMcpVersion,
		}, &mcp.ServerOptions{
			KeepAlive: 30 * time.Second,
			HasTools:  true,
		}),
	}

	server.addTools()

	return server
}

func (s *DaytonaFileSystemMCPServer) Start(ctx context.Context, transport mcp.Transport) error {
	return s.Server.Run(ctx, transport)
}

func (s *DaytonaFileSystemMCPServer) addTools() {
	mcp.AddTool(s.Server, getCreateFolderTool(), handleCreateFolder)
	mcp.AddTool(s.Server, getDeleteFileTool(), handleDeleteFile)
	mcp.AddTool(s.Server, getDownloadFileTool(), handleDownloadFile)
	mcp.AddTool(s.Server, getFileInfoTool(), handleFileInfo)
	mcp.AddTool(s.Server, getListFilesTool(), handleListFiles)
	mcp.AddTool(s.Server, getMoveFileTool(), handleMoveFile)
	mcp.AddTool(s.Server, getUploadFileTool(), handleUploadFile)
}
