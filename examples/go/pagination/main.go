// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
)

func main() {
	// Create a new Daytona client using environment variables
	// Set DAYTONA_API_KEY before running
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Paginate through sandboxes with labels
	log.Println("=== Example 1: Paginate Sandboxes with Labels ===")
	labels := map[string]string{
		"my-label": "my-value",
	}
	page := 2
	limit := 10

	sandboxList, err := client.List(ctx, labels, &page, &limit)
	if err != nil {
		log.Fatalf("Failed to list sandboxes: %v", err)
	}

	log.Printf("Total sandboxes: %d\n", sandboxList.Total)
	log.Printf("Page: %d, Limit: %d\n", page, limit)
	for _, sandbox := range sandboxList.Items {
		log.Printf("  - %s: %s\n", sandbox.ID, sandbox.State)
	}

	// Example 2: Paginate through snapshots
	log.Println("\n=== Example 2: Paginate Snapshots ===")
	snapshotPage := 2
	snapshotLimit := 10

	snapshotList, err := client.Snapshot.List(ctx, &snapshotPage, &snapshotLimit)
	if err != nil {
		log.Fatalf("Failed to list snapshots: %v", err)
	}

	log.Printf("Found %d snapshots\n", snapshotList.Total)
	log.Printf("Page: %d, Limit: %d\n", snapshotPage, snapshotLimit)
	for _, snapshot := range snapshotList.Items {
		log.Printf("  - %s (%s)\n", snapshot.Name, snapshot.ImageName)
	}

	log.Println("\n✓ All pagination examples completed successfully!")
}
