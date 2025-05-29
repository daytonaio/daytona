// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

func FileUpload(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return nil, err
	}

	sandboxId := ""
	if id, ok := request.Params.Arguments["id"]; ok && id != nil {
		if idStr, ok := id.(string); ok && idStr != "" {
			sandboxId = idStr
		}
	}

	if sandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	filePath := request.Params.Arguments["file_path"].(string)
	content := request.Params.Arguments["content"].(string)
	encoding := request.Params.Arguments["encoding"].(string)
	overwrite := request.Params.Arguments["overwrite"].(bool)

	// Get the sandbox using sandbox ID
	sandbox, _, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxId).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to get sandbox: %v", err)
	}

	if sandbox == nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no sandbox available")
	}

	// Check if file exists and handle overwrite
	if !overwrite {
		fileInfo, _, err := apiClient.ToolboxAPI.GetFileInfo(ctx, sandboxId).Path(filePath).Execute()
		if err == nil && fileInfo != nil {
			return &mcp.CallToolResult{IsError: true}, fmt.Errorf("file '%s' already exists and overwrite=false", filePath)
		}
	}

	// Prepare content based on encoding
	var binaryContent []byte
	if encoding == "base64" {
		var err error
		binaryContent, err = base64.StdEncoding.DecodeString(content)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, fmt.Errorf("invalid base64 encoding: %v", err)
		}
	} else {
		// Default is text encoding
		binaryContent = []byte(content)
	}

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(filePath)
	if parentDir != "" {
		_, err := apiClient.ToolboxAPI.CreateFolder(ctx, sandboxId).Path(parentDir).Mode("0755").Execute()
		if err != nil {
			log.Errorf("Error creating parent directory: %v", err)
			// Continue anyway as upload might handle this
		}
	}

	// Upload the file
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up temp file when done
	defer tempFile.Close()

	// Write content to temp file
	if _, err := tempFile.Write(binaryContent); err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error writing to temp file: %v", err)
	}

	// Reset file pointer to beginning
	if _, err := tempFile.Seek(0, 0); err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error seeking temp file: %v", err)
	}

	// Upload the file
	_, err = apiClient.ToolboxAPI.UploadFile(ctx, sandboxId).Path(filePath).File(tempFile).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error uploading file: %v", err)
	}

	// Get file info for size
	fileInfo, _, err := apiClient.ToolboxAPI.GetFileInfo(ctx, sandboxId).Path(filePath).Execute()
	if err != nil {
		log.Errorf("Error getting file info after upload: %v", err)

		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error getting file info after upload: %v", err)
	}

	fileSizeKB := float64(fileInfo.Size) / 1024
	log.Infof("File uploaded successfully: %s, size: %.2fKB", filePath, fileSizeKB)

	return mcp.NewToolResultText(fmt.Sprintf("File uploaded successfully: %s", filePath)), nil
}
