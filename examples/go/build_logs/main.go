// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/google/uuid"
)

func main() {
	// Create a new Daytona client using environment variables
	// Set DAYTONA_API_KEY before running
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Simple log streaming with direct output
	log.Println("=== Example 1: Simple Build Log Streaming ===")
	simpleLogStreaming(ctx, client)

	log.Println("\n✓ All build log streaming examples completed successfully!")
}

// Example 1: Simple log streaming with direct output
func simpleLogStreaming(ctx context.Context, client *daytona.Client) {
	log.Println("Creating sandbox with simple log streaming...")

	image := daytona.Base("python:3.12-slim").
		PipInstall([]string{"requests"}, options.WithFindLinks("https://pypi.org/simple")).
		Workdir("/app")

	name := fmt.Sprintf("simple-logs-sandbox-%s", uuid.New().String())
	params := types.ImageParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Name: name,
		},
		Image: image,
		Resources: &types.Resources{
			CPU:    1,
			Memory: 1,
		},
	}

	logChan := make(chan string, 100)

	go func() {
		for logLine := range logChan {
			fmt.Printf("[BUILD] %s\n", logLine)
		}
	}()

	sandbox, err := client.Create(ctx, params, options.WithTimeout(120*time.Second), options.WithLogChannel(logChan))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	log.Printf("✓ Sandbox created: %s (ID: %s)\n", sandbox.Name, sandbox.ID)

	// Clean up
	defer func() {
		log.Println("Cleaning up sandbox...")
		if err := sandbox.Delete(ctx); err != nil {
			log.Printf("Failed to delete sandbox: %v", err)
		}
	}()
}
