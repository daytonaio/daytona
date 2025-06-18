// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mcp

import (
	"github.com/daytonaio/daytona/cli/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type DaytonaMCPServer struct {
	server.MCPServer
}

func NewDaytonaMCPServer() *DaytonaMCPServer {
	s := &DaytonaMCPServer{}

	s.MCPServer = *server.NewMCPServer(
		"Daytona MCP Server",
		"0.1.0",
		server.WithRecovery(),
	)

	s.addTools()

	return s
}

func (s *DaytonaMCPServer) Start() error {
	return server.ServeStdio(&s.MCPServer)
}

func (s *DaytonaMCPServer) addTools() {
	s.AddTool(tools.GetCreateSandboxTool(), mcp.NewTypedToolHandler(tools.CreateSandbox))
	s.AddTool(tools.GetDestroySandboxTool(), mcp.NewTypedToolHandler(tools.DestroySandbox))
	s.AddTool(tools.GetDownloadFileTool(), mcp.NewTypedToolHandler(tools.DownloadCommand))
	s.AddTool(tools.GetFileUploadTool(), mcp.NewTypedToolHandler(tools.FileUpload))
	s.AddTool(tools.GetPreviewLinkTool(), mcp.NewTypedToolHandler(tools.PreviewLink))
	s.AddTool(tools.GetExecuteCommandTool(), mcp.NewTypedToolHandler(tools.ExecuteCommand))
	s.AddTool(tools.GetCreateFolderTool(), mcp.NewTypedToolHandler(tools.CreateFolder))
	s.AddTool(tools.GetGetFileInfoTool(), mcp.NewTypedToolHandler(tools.GetFileInfo))
	s.AddTool(tools.GetListFilesTool(), mcp.NewTypedToolHandler(tools.ListFiles))
	s.AddTool(tools.GetGitCloneTool(), mcp.NewTypedToolHandler(tools.GitClone))
	s.AddTool(tools.GetMoveFileTool(), mcp.NewTypedToolHandler(tools.MoveFile))
	s.AddTool(tools.GetDeleteFileTool(), mcp.NewTypedToolHandler(tools.DeleteFile))
}
