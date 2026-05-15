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

	// Create a sandbox
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
	log.Printf("✓ Sandbox created: %s (ID: %s, State: %s)\n", sandbox.Name, sandbox.ID, sandbox.State)

	// Set labels on the sandbox
	log.Println("\nSetting labels on sandbox...")
	labels := map[string]string{
		"public": "true",
		"env":    "development",
	}
	if err := sandbox.SetLabels(ctx, labels); err != nil {
		log.Fatalf("Failed to set labels: %v", err)
	}
	log.Println("✓ Labels set successfully")

	// Stop the sandbox
	log.Println("\nStopping sandbox...")
	if err := sandbox.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop sandbox: %v", err)
	}
	log.Println("✓ Sandbox stopped")

	// Get the sandbox info to verify state
	sandbox, err = client.Get(ctx, sandbox.ID)
	if err != nil {
		log.Fatalf("Failed to get sandbox: %v", err)
	}
	log.Printf("Sandbox state after stop: %s\n", sandbox.State)

	// Start the sandbox
	log.Println("\nStarting sandbox...")
	if err := sandbox.Start(ctx); err != nil {
		log.Fatalf("Failed to start sandbox: %v", err)
	}
	log.Println("✓ Sandbox started")

	// Get the sandbox info again
	log.Println("\nGetting existing sandbox...")
	existingSandbox, err := client.Get(ctx, sandbox.ID)
	if err != nil {
		log.Fatalf("Failed to get sandbox: %v", err)
	}
	log.Printf("✓ Got existing sandbox: %s (State: %s)\n", existingSandbox.Name, existingSandbox.State)

	// Execute a command to verify it's running
	log.Println("\nExecuting command on sandbox...")
	result, err := existingSandbox.Process.ExecuteCommand(ctx, "echo \"Hello World from exec!\"")
	if err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}

	if result.ExitCode != 0 {
		log.Printf("Command failed with exit code %d: %s\n", result.ExitCode, result.Result)
	} else {
		log.Printf("Command output: %s\n", result.Result)
	}

	// List all sandboxes (Go 1.23+ range-over-func)
	limit := 10
	sort := daytona.SandboxListSortFieldCreatedAt
	order := daytona.SandboxListSortDirectionDesc
	for sb, err := range client.ListSeq(ctx, &daytona.ListSandboxesQuery{
		Limit:  &limit,
		Labels: map[string]string{"env": "dev"},
		States: []daytona.SandboxState{daytona.SandboxStateStarted},
		Sort:   &sort,
		Order:  &order,
	}) {
		if err != nil {
			log.Fatalf("Failed to list sandboxes: %v", err)
		}
		log.Println(sb.ID)
	}

	// Delete the sandbox
	log.Println("\nDeleting sandbox...")
	if err := sandbox.Delete(ctx); err != nil {
		log.Fatalf("Failed to delete sandbox: %v", err)
	}
	log.Println("✓ Sandbox deleted")

	log.Println("\n✓ All lifecycle operations completed successfully!")
}
