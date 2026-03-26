// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/google/uuid"
)

// This example demonstrates creating a snapshot with log streaming
func exampleSnapshotWithLogs() {
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Create a snapshot from a simple base image with log streaming
	log.Println("Creating snapshot from base image with log streaming...")
	log.Println("=" + string(make([]byte, 60)) + "=")

	snapshotName := fmt.Sprintf("example-snapshot-with-logs-%s", uuid.New().String())

	// Define the snapshot creation parameters
	params := &types.CreateSnapshotParams{
		Name:  snapshotName,
		Image: "ubuntu:22.04",
		Resources: &types.Resources{
			CPU:    1,
			Memory: 1,
		},
	}

	// Create the snapshot and get the log channel
	snapshot, logChan, err := client.Snapshot.Create(ctx, params)
	if err != nil {
		log.Fatalf("Failed to create snapshot: %v", err)
	}

	// Read logs from the channel until it's closed
	for logLine := range logChan {
		fmt.Println(logLine)
	}

	log.Printf("\n✓ Snapshot created successfully!")
	log.Printf("  Name: %s\n", snapshot.Name)
	log.Printf("  ID: %s\n", snapshot.ID)
	log.Printf("  State: %s\n", snapshot.State)
	log.Printf("  Image: %s\n", snapshot.ImageName)

	// Optional: Clean up - delete the snapshot
	log.Println("\nCleaning up - deleting snapshot...")
	if err := client.Snapshot.Delete(ctx, snapshot); err != nil {
		log.Printf("Warning: Failed to delete snapshot: %v", err)
	} else {
		log.Println("✓ Snapshot deleted successfully!")
	}
}

// This example demonstrates creating a snapshot with a custom Dockerfile
func exampleSnapshotWithCustomImage() {
	// Create a new Daytona client
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	log.Println("\n\nCreating snapshot with custom image (Python + packages)...")
	log.Println("=" + string(make([]byte, 60)) + "=")

	// Build a custom image with Python and some packages
	image := daytona.Base("python:3.11-slim").
		Run("apt-get update && apt-get install -y git curl").
		Run("pip install --no-cache-dir requests numpy pandas")

	name := fmt.Sprintf("example-python-snapshot-%s", uuid.New().String())
	params := &types.CreateSnapshotParams{
		Name:  name,
		Image: image,
		Resources: &types.Resources{
			CPU:    1,
			Memory: 1,
		},
	}

	// Create the snapshot and get the log channel
	snapshot, logChan, err := client.Snapshot.Create(ctx, params)
	if err != nil {
		log.Fatalf("Failed to create snapshot: %v", err)
	}

	// Read logs from the channel until it's closed
	for logLine := range logChan {
		fmt.Println("This LOG is from the channel: " + logLine)
	}

	log.Printf("\n✓ Custom snapshot created successfully!")
	log.Printf("  Name: %s\n", snapshot.Name)
	log.Printf("  ID: %s\n", snapshot.ID)
	log.Printf("  State: %s\n", snapshot.State)

	// Optional: Clean up
	log.Println("\nCleaning up - deleting snapshot...")
	if err := client.Snapshot.Delete(ctx, snapshot); err != nil {
		log.Printf("Warning: Failed to delete snapshot: %v", err)
	} else {
		log.Println("✓ Snapshot deleted successfully!")
	}
}

func main() {
	log.Println("Snapshot with Log Streaming Examples")
	log.Println("=====================================")

	exampleSnapshotWithLogs()

	exampleSnapshotWithCustomImage()

	log.Println("\n\n✓ All examples completed successfully!")
}
