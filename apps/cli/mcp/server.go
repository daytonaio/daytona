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
	createSandboxTool := mcp.NewTool("create_sandbox",
		mcp.WithDescription("Create a new sandbox with Daytona"),
		mcp.WithString("id", mcp.Description("If a sandbox ID is provided it is first checked if it exists and is running, if so, the existing sandbox will be used. However, a model is not able to provide custom sandbox ID but only the ones Daytona commands return and should always leave ID field empty if the intention is to create a new sandbox.")),
		mcp.WithString("target", mcp.DefaultString("us"), mcp.Description("Target region of the sandbox.")),
		mcp.WithString("snapshot", mcp.Description("Snapshot of the sandbox (don't specify any if not explicitly instructed from user).")),
		mcp.WithString("auto_stop_interval", mcp.DefaultString("15"), mcp.Description("Auto-stop interval in minutes (0 means disabled) for the sandbox.")),
		mcp.WithString("auto_archive_interval", mcp.DefaultString("10080"), mcp.Description("Auto-archive interval in minutes (0 means the maximum interval will be used) for the sandbox.")),
	)

	s.AddTool(createSandboxTool, tools.CreateSandbox)

	destroySandboxTool := mcp.NewTool("destroy_sandbox",
		mcp.WithDescription("Destroy a sandbox with Daytona"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to destroy.")),
	)

	s.AddTool(destroySandboxTool, tools.DestroySandbox)

	downloadFileTool := mcp.NewTool("download_file",
		mcp.WithDescription("Download a file from the Daytona sandbox. Returns the file content either as text or as a base64 encoded image. Handles special cases like matplotlib plots stored as JSON with embedded base64 images."),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file to download.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to download the file from.")),
	)

	s.AddTool(downloadFileTool, tools.DownloadCommand)

	uploadFileTool := mcp.NewTool("file_upload",
		mcp.WithDescription("Upload files to the Daytona sandbox from text or base64-encoded binary content. Creates necessary parent directories automatically and verifies successful writes. Files persist during the session and have appropriate permissions for further tool operations. Supports overwrite controls and maintains original file formats."),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file to upload. Files should always be uploaded to the /tmp directory if user doesn't specify otherwise.")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content of the file to upload.")),
		mcp.WithString("encoding", mcp.Required(), mcp.Description("Encoding of the file to upload.")),
		mcp.WithBoolean("overwrite", mcp.Required(), mcp.Description("Overwrite the file if it already exists.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to upload the file to.")),
	)

	s.AddTool(uploadFileTool, tools.FileUpload)

	previewLinkTool := mcp.NewTool("preview_link",
		mcp.WithDescription("Generate accessible preview URLs for web applications running in the Daytona sandbox. Creates a secure tunnel to expose local ports externally without configuration. Validates if a server is actually running on the specified port and provides diagnostic information for troubleshooting. Supports custom descriptions and metadata for better organization of multiple services."),
		mcp.WithString("port", mcp.Required(), mcp.Description("Port to expose.")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Description of the service.")),
		mcp.WithBoolean("check_server", mcp.Required(), mcp.Description("Check if a server is running on the specified port.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to generate the preview link for.")),
	)

	s.AddTool(previewLinkTool, tools.PreviewLink)

	executeCommandTool := mcp.NewTool("execute_command",
		mcp.WithDescription("Execute shell commands in the ephemeral Daytona Linux environment. Returns full stdout and stderr output with exit codes. Commands have sandbox user permissions and can install packages, modify files, and interact with running services. Always use /tmp directory. Use verbose flags where available for better output."),
		mcp.WithString("command", mcp.Required(), mcp.Description("Command to execute.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to execute the command in.")),
	)

	s.AddTool(executeCommandTool, tools.ExecuteCommand)

	createFolderTool := mcp.NewTool("create_folder",
		mcp.WithDescription("Create a new folder in the Daytona sandbox."),
		mcp.WithString("folder_path", mcp.Required(), mcp.Description("Path to the folder to create.")),
		mcp.WithString("mode", mcp.Description("Mode of the folder to create (defaults to 0755).")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to create the folder in.")),
	)

	s.AddTool(createFolderTool, tools.CreateFolder)

	getFileInfoTool := mcp.NewTool("get_file_info",
		mcp.WithDescription("Get information about a file in the Daytona sandbox."),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file to get information about.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to get the file information from.")),
	)

	s.AddTool(getFileInfoTool, tools.GetFileInfo)

	listFilesTool := mcp.NewTool("list_files",
		mcp.WithDescription("List files in a directory in the Daytona sandbox."),
		mcp.WithString("path", mcp.Description("Path to the directory to list files from (defaults to current directory).")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to list the files from.")),
	)

	s.AddTool(listFilesTool, tools.ListFiles)

	gitCloneTool := mcp.NewTool("git_clone",
		mcp.WithDescription("Clone a Git repository into the Daytona sandbox."),
		mcp.WithString("url", mcp.Required(), mcp.Description("URL of the Git repository to clone.")),
		mcp.WithString("path", mcp.Description("Directory to clone the repository into (defaults to current directory).")),
		mcp.WithString("branch", mcp.Description("Branch to clone.")),
		mcp.WithString("commit_id", mcp.Description("Commit ID to clone.")),
		mcp.WithString("username", mcp.Description("Username to clone the repository with.")),
		mcp.WithString("password", mcp.Description("Password to clone the repository with.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to clone the repository in.")),
	)

	s.AddTool(gitCloneTool, tools.GitClone)

	moveFileTool := mcp.NewTool("move_file",
		mcp.WithDescription("Move or rename a file in the Daytona sandbox."),
		mcp.WithString("source_path", mcp.Required(), mcp.Description("Source path of the file to move.")),
		mcp.WithString("dest_path", mcp.Required(), mcp.Description("Destination path where to move the file.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to move the file in.")),
	)

	s.AddTool(moveFileTool, tools.MoveFile)

	deleteFileTool := mcp.NewTool("delete_file",
		mcp.WithDescription("Delete a file or directory in the Daytona sandbox."),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file or directory to delete.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to delete the file in.")),
	)

	s.AddTool(deleteFileTool, tools.DeleteFile)
}
