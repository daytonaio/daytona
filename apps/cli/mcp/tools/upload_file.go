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

type FileUploadArgs struct {
	Id        *string `json:"id,omitempty"`
	FilePath  *string `json:"filePath,omitempty"`
	Content   *string `json:"content,omitempty"`
	Encoding  *string `json:"encoding,omitempty"`
	Overwrite *bool   `json:"overwrite,omitempty"`
}

func GetFileUploadTool() mcp.Tool {
	return mcp.NewTool("file_upload",
		mcp.WithDescription("Upload files to the Daytona sandbox from text or base64-encoded binary content. Creates necessary parent directories automatically and verifies successful writes. Files persist during the session and have appropriate permissions for further tool operations. Supports overwrite controls and maintains original file formats."),
		mcp.WithString("filePath", mcp.Required(), mcp.Description("Path to the file to upload. Files should always be uploaded to the /tmp directory if user doesn't specify otherwise.")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content of the file to upload.")),
		mcp.WithString("encoding", mcp.Required(), mcp.Description("Encoding of the file to upload.")),
		mcp.WithBoolean("overwrite", mcp.Required(), mcp.Description("Overwrite the file if it already exists.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to upload the file to.")),
	)
}

func FileUpload(ctx context.Context, request mcp.CallToolRequest, args FileUploadArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return nil, err
	}

	if args.Id == nil || *args.Id == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	if args.FilePath == nil || *args.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("filePath parameter is required")
	}

	if args.Content == nil || *args.Content == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("content parameter is required")
	}

	if args.Encoding == nil || *args.Encoding == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("encoding parameter is required")
	}

	overwrite := false
	if args.Overwrite != nil && *args.Overwrite {
		overwrite = *args.Overwrite
	}

	// Get the sandbox using sandbox ID
	sandbox, _, err := apiClient.SandboxAPI.GetSandbox(ctx, *args.Id).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to get sandbox: %v", err)
	}

	if sandbox == nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no sandbox available")
	}

	// Check if file exists and handle overwrite
	if !overwrite {
		fileInfo, _, err := apiClient.ToolboxAPI.GetFileInfo(ctx, *args.Id).Path(*args.FilePath).Execute()
		if err == nil && fileInfo != nil {
			return &mcp.CallToolResult{IsError: true}, fmt.Errorf("file '%s' already exists and overwrite=false", *args.FilePath)
		}
	}

	// Prepare content based on encoding
	var binaryContent []byte
	if *args.Encoding == "base64" {
		var err error
		binaryContent, err = base64.StdEncoding.DecodeString(*args.Content)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, fmt.Errorf("invalid base64 encoding: %v", err)
		}
	} else {
		// Default is text encoding
		binaryContent = []byte(*args.Content)
	}

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(*args.FilePath)
	if parentDir != "" {
		_, err := apiClient.ToolboxAPI.CreateFolder(ctx, *args.Id).Path(parentDir).Mode("0755").Execute()
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
	_, err = apiClient.ToolboxAPI.UploadFile(ctx, *args.Id).Path(*args.FilePath).File(tempFile).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error uploading file: %v", err)
	}

	// Get file info for size
	fileInfo, _, err := apiClient.ToolboxAPI.GetFileInfo(ctx, *args.Id).Path(*args.FilePath).Execute()
	if err != nil {
		log.Errorf("Error getting file info after upload: %v", err)

		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error getting file info after upload: %v", err)
	}

	fileSizeKB := float64(fileInfo.Size) / 1024
	log.Infof("File uploaded successfully: %s, size: %.2fKB", *args.FilePath, fileSizeKB)

	return mcp.NewToolResultText(fmt.Sprintf("File uploaded successfully: %s", *args.FilePath)), nil
}
