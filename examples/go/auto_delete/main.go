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

	// Example 1: Auto-delete is disabled by default
	log.Println("=== Example 1: Default auto-delete interval (disabled) ===")
	params1 := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language: types.CodeLanguagePython,
		},
	}

	sandbox1, err := client.Create(ctx, params1, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Default auto-delete interval: %d (disabled)\n", sandbox1.AutoDeleteInterval)

	// Example 2: Auto-delete after the Sandbox has been stopped for 1 hour
	log.Println("\n=== Example 2: Set auto-delete to 60 minutes after stop ===")
	interval := 60
	if err := sandbox1.SetAutoDeleteInterval(ctx, &interval); err != nil {
		log.Fatalf("Failed to set auto-delete interval: %v", err)
	}

	// Refresh sandbox info
	sandbox1, err = client.Get(ctx, sandbox1.ID)
	if err != nil {
		log.Fatalf("Failed to get sandbox: %v", err)
	}
	log.Printf("Updated auto-delete interval: %d minutes\n", sandbox1.AutoDeleteInterval)

	// Example 3: Delete immediately upon stopping
	log.Println("\n=== Example 3: Delete immediately upon stopping (0 minutes) ===")
	immediateInterval := 0
	if err := sandbox1.SetAutoDeleteInterval(ctx, &immediateInterval); err != nil {
		log.Fatalf("Failed to set auto-delete interval: %v", err)
	}

	// Refresh sandbox info
	sandbox1, err = client.Get(ctx, sandbox1.ID)
	if err != nil {
		log.Fatalf("Failed to get sandbox: %v", err)
	}
	log.Printf("Updated auto-delete interval: %d (immediate)\n", sandbox1.AutoDeleteInterval)

	// Example 4: Disable auto-delete
	log.Println("\n=== Example 4: Disable auto-delete (-1) ===")
	disableInterval := -1
	if err := sandbox1.SetAutoDeleteInterval(ctx, &disableInterval); err != nil {
		log.Fatalf("Failed to set auto-delete interval: %v", err)
	}

	// Refresh sandbox info
	sandbox1, err = client.Get(ctx, sandbox1.ID)
	if err != nil {
		log.Fatalf("Failed to get sandbox: %v", err)
	}
	log.Printf("Updated auto-delete interval: %d (disabled)\n", sandbox1.AutoDeleteInterval)

	// Clean up first sandbox
	if err := sandbox1.Delete(ctx); err != nil {
		log.Printf("Failed to delete sandbox: %v", err)
	}

	// Example 5: Auto-delete after the Sandbox has been stopped for 1 day
	log.Println("\n=== Example 5: Sandbox with 1 day auto-delete interval ===")
	oneDayInterval := 1440 // 24 hours * 60 minutes
	params2 := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language:           types.CodeLanguagePython,
			AutoDeleteInterval: &oneDayInterval,
		},
	}

	sandbox2, err := client.Create(ctx, params2, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Auto-delete interval: %d minutes (1 day)\n", sandbox2.AutoDeleteInterval)

	// Clean up second sandbox
	if err := sandbox2.Delete(ctx); err != nil {
		log.Printf("Failed to delete sandbox: %v", err)
	}

	log.Println("\n✓ All auto-delete examples completed successfully!")
}
