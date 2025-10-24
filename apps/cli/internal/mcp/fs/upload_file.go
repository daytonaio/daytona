// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	"github.com/invopop/jsonschema"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type FileUploadInput struct {
	SandboxId *string `json:"sandboxId,omitempty" jsonschema:"required,description=ID of the sandbox to upload the file to."`
	FilePath  *string `json:"filePath,omitempty" jsonschema:"required,description=Path to the file to upload. Files should always be uploaded to the /tmp directory if user doesn't specify otherwise."`
	Content   *string `json:"content,omitempty" jsonschema:"required,description=Content of the file to upload."`
	Encoding  *string `json:"encoding,omitempty" jsonschema:"default=text,description=Encoding of the file to upload."`
	Overwrite *bool   `json:"overwrite,omitempty" jsonschema:"default=false,description=Overwrite the file if it already exists."`
}

type FileUploadOutput struct {
	Message string `json:"message" jsonschema:"description=Message indicating the successful upload of the file."`
}

func getUploadFileTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "file_upload",
		Title:        "Upload File",
		Description:  "Upload files to the Daytona sandbox from text or base64-encoded binary content. Creates necessary parent directories automatically and verifies successful writes. Files persist during the session and have appropriate permissions for further tool operations. Supports overwrite controls and maintains original file formats.",
		InputSchema:  jsonschema.Reflect(FileUploadInput{}),
		OutputSchema: jsonschema.Reflect(FileUploadOutput{}),
	}
}

func handleUploadFile(ctx context.Context, request *mcp.CallToolRequest, input *FileUploadInput) (*mcp.CallToolResult, *FileUploadOutput, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	if input.SandboxId == nil || *input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	if input.FilePath == nil || *input.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("filePath parameter is required")
	}

	if input.Content == nil || *input.Content == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("content parameter is required")
	}

	if input.Encoding == nil {
		defaultEncoding := "text"
		input.Encoding = &defaultEncoding
	}

	overwrite := false
	if input.Overwrite != nil && *input.Overwrite {
		overwrite = *input.Overwrite
	}

	// Get the sandbox using sandbox ID
	sandbox, _, err := apiClient.SandboxAPI.GetSandbox(ctx, *input.SandboxId).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to get sandbox: %v", err)
	}

	if sandbox == nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("no sandbox available")
	}

	// Check if file exists and handle overwrite
	if !overwrite {
		fileInfo, _, err := apiClient.ToolboxAPI.GetFileInfo(ctx, *input.SandboxId).Path(*input.FilePath).Execute()
		if err == nil && fileInfo != nil {
			return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("file '%s' already exists and overwrite=false", *input.FilePath)
		}
	}

	// Prepare content based on encoding
	var binaryContent []byte
	if *input.Encoding == "base64" {
		var err error
		binaryContent, err = base64.StdEncoding.DecodeString(*input.Content)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("invalid base64 encoding: %v", err)
		}
	} else {
		// Default is text encoding
		binaryContent = []byte(*input.Content)
	}

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(*input.FilePath)
	if parentDir != "" {
		_, err := apiClient.ToolboxAPI.CreateFolder(ctx, *input.SandboxId).Path(parentDir).Mode("0755").Execute()
		if err != nil {
			log.Errorf("Error creating parent directory: %v", err)
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
	_, err = apiClient.ToolboxAPI.UploadFile(ctx, *input.SandboxId).Path(*input.FilePath).File(tempFile).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error uploading file: %v", err)
	}

	// Get file info for size
	fileInfo, _, err := apiClient.ToolboxAPI.GetFileInfo(ctx, *input.SandboxId).Path(*input.FilePath).Execute()
	if err != nil {
		log.Errorf("Error getting file info after upload: %v", err)

		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error getting file info after upload: %v", err)
	}

	fileSizeKB := float64(fileInfo.Size) / 1024
	log.Infof("File uploaded successfully: %s, size: %.2fKB", *input.FilePath, fileSizeKB)

	return &mcp.CallToolResult{
			IsError: false,
		}, &FileUploadOutput{
			Message: fmt.Sprintf("File uploaded successfully: %s", *input.FilePath),
		}, nil
}
