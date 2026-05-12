// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"strings"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
	// Create a new Daytona client using environment variables
	// Set DAYTONA_API_KEY before running
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Create a sandbox with Python
	log.Println("Creating sandbox...")
	params := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language: types.CodeLanguagePython,
		},
	}

	sandbox, err := client.Create(ctx, params, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	log.Printf("✓ Created sandbox: %s (ID: %s)\n", sandbox.Name, sandbox.ID)
	defer func() {
		log.Println("\nCleaning up...")
		if err := sandbox.Delete(ctx); err != nil {
			log.Printf("Failed to delete sandbox: %v", err)
		} else {
			log.Println("✓ Sandbox deleted")
		}
	}()

	// File system operations
	log.Println("\nPerforming file operations...")

	// Create a test file
	testContent := []byte("Hello, Daytona!\nThis is a test file.")
	testPath := "/tmp/test.txt"

	if err := sandbox.FileSystem.UploadFile(ctx, testContent, testPath); err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}
	log.Printf("✓ Uploaded file to %s\n", testPath)

	// Download the file
	downloadedContent, err := sandbox.FileSystem.DownloadFile(ctx, testPath, nil)
	if err != nil {
		log.Fatalf("Failed to download file: %v", err)
	}
	log.Printf("✓ Downloaded file content: %s\n", string(downloadedContent))

	// Stream upload — push any io.Reader straight to the sandbox without
	// buffering the whole payload, with live progress reporting.
	streamedPath := "/tmp/streamed.bin"
	generatedPayload := []byte(strings.Repeat("streamed-upload-content-", 2048)) // ~48 KB
	uploadErr := sandbox.FileSystem.UploadFileStream(
		ctx,
		bytes.NewReader(generatedPayload),
		streamedPath,
		daytona.WithUploadProgress(func(p daytona.UploadProgress) {
			log.Printf("  uploaded %d / %d bytes\n", p.BytesSent, len(generatedPayload))
		}),
	)
	if uploadErr != nil {
		log.Fatalf("Failed to stream upload: %v", uploadErr)
	}
	log.Printf("✓ Streamed upload to %s (%d bytes)\n", streamedPath, len(generatedPayload))

	// Stream download — process file content as chunks arrive, with progress.
	// Cancel a long-running transfer by cancelling the context.
	stream, err := sandbox.FileSystem.DownloadFileStream(
		ctx,
		testPath,
		daytona.WithDownloadProgress(func(p daytona.DownloadProgress) {
			log.Printf("  downloaded %d / %d bytes\n", p.BytesReceived, p.TotalBytes)
		}),
	)
	if err != nil {
		log.Fatalf("Failed to stream download file: %v", err)
	}
	streamedContent, err := io.ReadAll(stream)
	stream.Close()
	if err != nil {
		log.Fatalf("Failed to read stream: %v", err)
	}
	log.Printf("✓ Streamed file content: %s\n", string(streamedContent))

	// Create a folder
	folderPath := "/tmp/test-folder"
	if err := sandbox.FileSystem.CreateFolder(ctx, folderPath); err != nil {
		log.Fatalf("Failed to create folder: %v", err)
	}
	log.Printf("✓ Created folder at %s\n", folderPath)

	// List files in /tmp
	files, err := sandbox.FileSystem.ListFiles(ctx, "/tmp")
	if err != nil {
		log.Fatalf("Failed to list files: %v", err)
	}
	log.Printf("\nFiles in /tmp:\n")
	for _, file := range files {
		log.Printf("  - %s (%d bytes)\n", file.Name, file.Size)
	}

	log.Println("\n✓ All file operations completed successfully!")
}
