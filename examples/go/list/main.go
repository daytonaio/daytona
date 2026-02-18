// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
)

func main() {
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	limit := 2
	states := []apiclient.SandboxState{
		apiclient.SANDBOXSTATE_STARTED,
		apiclient.SANDBOXSTATE_STOPPED,
	}

	page1, err := client.ListV2(ctx, &daytona.ListSandboxesParams{
		Limit:  &limit,
		States: states,
	})
	if err != nil {
		log.Fatalf("Failed to list sandboxes: %v", err)
	}

	for _, sandbox := range page1.Items {
		log.Printf("%s: %s", sandbox.ID, sandbox.State)
	}

	if page1.NextCursor != nil {
		page2, err := client.ListV2(ctx, &daytona.ListSandboxesParams{
			Cursor: page1.NextCursor,
			Limit:  &limit,
			States: states,
		})
		if err != nil {
			log.Fatalf("Failed to list sandboxes: %v", err)
		}

		for _, sandbox := range page2.Items {
			log.Printf("%s: %s", sandbox.ID, sandbox.State)
		}
	}
}
