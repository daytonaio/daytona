// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
)

func main() {
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Iterate through sandboxes
	limit := 10
	sort := "createdAt"
	order := "desc"
	iter := client.List(ctx, &daytona.ListSandboxesQuery{
		Limit:  &limit,
		Labels: map[string]string{"env": "dev"},
		States: []string{"started"},
		Sort:   &sort,
		Order:  &order,
	})
	defer iter.Close()
	for iter.Next() {
		sandbox := iter.Value()
		fmt.Println(sandbox.ID)
	}
	if err := iter.Err(); err != nil {
		log.Fatalf("Failed to list sandboxes: %v", err)
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
