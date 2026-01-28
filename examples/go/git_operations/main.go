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

	log.Printf("✓ Created sandbox: %s (ID: %s)\n", sandbox.Name, sandbox.ID)
	defer func() {
		log.Println("\nCleaning up...")
		if err := sandbox.Delete(ctx); err != nil {
			log.Printf("Failed to delete sandbox: %v", err)
		} else {
			log.Println("✓ Sandbox deleted")
		}
	}()

	// Git operations example
	log.Println("\nGit operations example...")
	repoPath := "/tmp/test-repo"

	// Create directory for repo
	if err := sandbox.FileSystem.CreateFolder(ctx, repoPath); err != nil {
		log.Fatalf("Failed to create repo directory: %v", err)
	}

	// Clone a public repository
	repoURL := "https://github.com/daytonaio/daytona.git"
	log.Printf("Cloning %s...\n", repoURL)
	if err := sandbox.Git.Clone(ctx, repoURL, repoPath); err != nil {
		log.Fatalf("Failed to clone repository: %v", err)
	}
	log.Println("✓ Repository cloned")

	// Example with options:
	// err := sandbox.Git.Clone(ctx, repoURL, repoPath,
	// 	options.WithBranch("main"),
	// 	options.WithUsername("user"),
	// 	options.WithPassword("pass"),
	// )

	// Get status
	status, err := sandbox.Git.Status(ctx, repoPath)
	if err != nil {
		log.Fatalf("Failed to get git status: %v", err)
	}
	log.Printf("Current branch: %s\n", status.CurrentBranch)

	// List branches
	branches, err := sandbox.Git.Branches(ctx, repoPath)
	if err != nil {
		log.Fatalf("Failed to list branches: %v", err)
	}
	log.Printf("Branches: %v\n", branches)

	log.Println("\n✓ All git operations completed successfully!")
}
