// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
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

	// Example 1: Default interval
	log.Println("=== Example 1: Default auto-archive interval ===")
	params1 := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language: types.CodeLanguagePython,
		},
	}

	sandbox1, err := client.Create(ctx, params1, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Default auto-archive interval: %d minutes\n", sandbox1.AutoArchiveInterval)

	// Example 2: Set interval to 1 hour
	log.Println("\n=== Example 2: Update auto-archive interval to 60 minutes ===")
	interval := 60
	if err := sandbox1.SetAutoArchiveInterval(ctx, &interval); err != nil {
		log.Fatalf("Failed to set auto-archive interval: %v", err)
	}

	// Refresh sandbox info to see the updated interval
	sandbox1, err = client.Get(ctx, sandbox1.ID)
	if err != nil {
		log.Fatalf("Failed to get sandbox: %v", err)
	}
	log.Printf("Updated auto-archive interval: %d minutes\n", sandbox1.AutoArchiveInterval)

	// Clean up first sandbox
	if err := sandbox1.Delete(ctx); err != nil {
		log.Printf("Failed to delete sandbox: %v", err)
	}

	// Example 3: Max interval
	log.Println("\n=== Example 3: Sandbox with max interval (never archive) ===")
	maxInterval := 0
	params2 := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language:            types.CodeLanguagePython,
			AutoArchiveInterval: &maxInterval,
		},
	}

	sandbox2, err := client.Create(ctx, params2, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Auto-archive interval: %d (never archive)\n", sandbox2.AutoArchiveInterval)

	// Clean up second sandbox
	if err := sandbox2.Delete(ctx); err != nil {
		log.Printf("Failed to delete sandbox: %v", err)
	}

	// Example 4: 1 day interval
	log.Println("\n=== Example 4: Sandbox with 1 day auto-archive interval ===")
	oneDayInterval := 1440 // 24 hours * 60 minutes
	params3 := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language:            types.CodeLanguagePython,
			AutoArchiveInterval: &oneDayInterval,
		},
	}

	sandbox3, err := client.Create(ctx, params3, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Auto-archive interval: %d minutes (1 day)\n", sandbox3.AutoArchiveInterval)

	// Clean up third sandbox
	if err := sandbox3.Delete(ctx); err != nil {
		log.Printf("Failed to delete sandbox: %v", err)
	}

	log.Println("\n✓ All auto-archive examples completed successfully!")
}
