// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/daytonaio/mcp/internal/apiclient"
	"github.com/daytonaio/mcp/internal/common"
	"github.com/daytonaio/mcp/internal/constants"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type FileUploadInput struct {
	SandboxId string `json:"sandboxId" jsonschema:"ID of the sandbox to upload the file to."`
	FilePath  string `json:"filePath" jsonschema:"Path to the file to upload. Files should always be uploaded to the /tmp directory if user doesn't specify otherwise."`
	Content   string `json:"content" jsonschema:"Content of the file to upload."`
	Encoding  string `json:"encoding,omitempty" jsonschema:"Encoding of the file to upload."`
	Overwrite bool   `json:"overwrite,omitempty" jsonschema:"Overwrite the file if it already exists."`
}

type FileUploadOutput struct {
	Message string `json:"message,omitempty" jsonschema:"Message indicating the successful upload of the file."`
}

func (s *DaytonaFileSystemMCPServer) getUploadFileTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "file_upload",
		Title:       "Upload File",
		Description: "Upload files to the Daytona sandbox from text or base64-encoded binary content. Creates necessary parent directories automatically and verifies successful writes. Files persist during the session and have appropriate permissions for further tool operations. Supports overwrite controls and maintains original file formats.",
	}
}

func (s *DaytonaFileSystemMCPServer) handleUploadFile(ctx context.Context, request *mcp.CallToolRequest, input *FileUploadInput) (*mcp.CallToolResult, *FileUploadOutput, error) {
	if input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	if input.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("filePath parameter is required")
	}

	if input.Content == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("content parameter is required")
	}

	if input.Encoding == "" {
		input.Encoding = "text"
	}

	sandbox, stop, err := common.GetSandbox(ctx, s.apiClient, &input.SandboxId)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to get sandbox: %v", err)
	}
	defer stop()

	proxyUrl, err := apiclient.ExtractProxyUrl(ctx, s.apiClient)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error extracting proxy URL: %v", err)
	}

	toolboxApiClient := apiclient.NewToolboxApiClient(constants.DAYTONA_FS_MCP_SOURCE, sandbox.Id, proxyUrl, request.Extra.Header)

	// Check if file exists and handle overwrite
	if !input.Overwrite {
		fileInfo, _, err := toolboxApiClient.FileSystemAPI.GetFileInfo(ctx).Path(input.FilePath).Execute()
		if err == nil && fileInfo != nil {
			return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("file '%s' already exists. Set overwrite to to overwrite the file", input.FilePath)
		}
	}

	// Prepare content based on encoding
	var binaryContent []byte
	if input.Encoding == "base64" {
		var err error
		binaryContent, err = base64.StdEncoding.DecodeString(input.Content)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("invalid base64 encoding: %v", err)
		}
	} else {
		// Default is text encoding
		binaryContent = []byte(input.Content)
	}

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(input.FilePath)
	if parentDir != "" {
		_, err := toolboxApiClient.FileSystemAPI.CreateFolder(ctx).Path(parentDir).Mode("0755").Execute()
		if err != nil {
			slog.Error("Error creating parent directory", "error", err)
			// Continue anyway as upload might handle this
		}
	}

	// Upload the file
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up temp file when done
	defer tempFile.Close()

	// Write content to temp file
	if _, err := tempFile.Write(binaryContent); err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error writing to temp file: %v", err)
	}

	// Reset file pointer to beginning
	if _, err := tempFile.Seek(0, 0); err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error seeking temp file: %v", err)
	}

	// Upload the file
	_, _, err = toolboxApiClient.FileSystemAPI.UploadFile(ctx).Path(input.FilePath).File(tempFile).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error uploading file: %v", err)
	}

	// Get file info for size
	fileInfo, _, err := toolboxApiClient.FileSystemAPI.GetFileInfo(ctx).Path(input.FilePath).Execute()
	if err != nil {
		slog.Error("Error getting file info after upload", "error", err)

		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error getting file info after upload: %v", err)
	}

	fileSizeKB := float64(fileInfo.Size) / 1024
	slog.Info("File uploaded successfully", "file_path", input.FilePath, "size", fileSizeKB)

	return &mcp.CallToolResult{
			IsError: false,
		}, &FileUploadOutput{
			Message: fmt.Sprintf("File uploaded successfully: %s", input.FilePath),
		}, nil
}
