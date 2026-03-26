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

	// Example 1: Default network settings
	log.Println("=== Example 1: Default network settings ===")
	params1 := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language: types.CodeLanguagePython,
		},
	}

	sandbox1, err := client.Create(ctx, params1, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Network Block All: %v\n", sandbox1.NetworkBlockAll)
	log.Printf("Network Allow List: %s\n", getNetworkAllowList(sandbox1))

	// Clean up
	if err := sandbox1.Delete(ctx); err != nil {
		log.Printf("Failed to delete sandbox: %v", err)
	}

	// Example 2: Block all network access
	log.Println("\n=== Example 2: Block all network access ===")
	blockAll := true
	params2 := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language:        types.CodeLanguagePython,
			NetworkBlockAll: blockAll,
		},
	}

	sandbox2, err := client.Create(ctx, params2, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Network Block All: %v\n", sandbox2.NetworkBlockAll)
	log.Printf("Network Allow List: %s\n", getNetworkAllowList(sandbox2))

	// Clean up
	if err := sandbox2.Delete(ctx); err != nil {
		log.Printf("Failed to delete sandbox: %v", err)
	}

	// Example 3: Explicitly allow list of network addresses
	log.Println("\n=== Example 3: Allow list of network addresses ===")
	allowList := "192.168.1.0/16,10.0.0.0/24"
	params3 := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language:         types.CodeLanguagePython,
			NetworkAllowList: &allowList,
		},
	}

	sandbox3, err := client.Create(ctx, params3, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Network Block All: %v\n", sandbox3.NetworkBlockAll)
	log.Printf("Network Allow List: %s\n", getNetworkAllowList(sandbox3))

	// Clean up
	if err := sandbox3.Delete(ctx); err != nil {
		log.Printf("Failed to delete sandbox: %v", err)
	}

	log.Println("\n✓ All network settings examples completed successfully!")
}

func getNetworkAllowList(sandbox *daytona.Sandbox) string {
	if sandbox.NetworkAllowList != nil {
		return *sandbox.NetworkAllowList
	}
	return "(not set)"
}
