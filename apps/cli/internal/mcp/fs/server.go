// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"time"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewDaytonaFileSystemMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "Daytona FileSystem MCP Server",
		Version: internal.FsMcpVersion,
	}, &mcp.ServerOptions{
		KeepAlive: 30 * time.Second,
		HasTools:  true,
	})

	addTools(server)

	return server
}

func addTools(server *mcp.Server) {
	mcp.AddTool(server, getCreateFolderTool(), handleCreateFolder)
	mcp.AddTool(server, getDeleteFileTool(), handleDeleteFile)
	mcp.AddTool(server, getDownloadFileTool(), handleDownloadFile)
	mcp.AddTool(server, getFileInfoTool(), handleFileInfo)
	mcp.AddTool(server, getListFilesTool(), handleListFiles)
	mcp.AddTool(server, getMoveFileTool(), handleMoveFile)
	mcp.AddTool(server, getUploadFileTool(), handleUploadFile)
}
