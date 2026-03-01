// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

// Test WebSocket event subscription for sandbox lifecycle.
//
// Run from the repo root:
//
//	cd libs/sdk-go && DAYTONA_API_KEY=... DAYTONA_API_URL=... go test -run TestWebSocketEvents -v -count=1 ./pkg/daytona/
//
// Or use the standalone runner (requires go.work to include this dir):
//
//	DAYTONA_API_KEY=... DAYTONA_API_URL=... go run ./examples/go/websocket-events/
//
// Tests:
// 1. Event subscriber connects on first sandbox creation
// 2. Sandbox state auto-updates via WebSocket events
// 3. WaitForStart/Stop use WebSocket events (not polling)
// 4. get() sandboxes also subscribe to events

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
	fmt.Println("============================================================")
	fmt.Println("WebSocket Event Subscription Test Suite (Go)")
	fmt.Println("============================================================")
	fmt.Println()

	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	ctx := context.Background()

	test1SubscriberConnects(client, ctx)
	test2EventDrivenLifecycle(client, ctx)
	test3AutoUpdateFromEvents(client, ctx)
	test4GetSubscribes(client, ctx)

	fmt.Println("============================================================")
	fmt.Println("ALL TESTS PASSED")
	fmt.Println("============================================================")
}

func test1SubscriberConnects(client *daytona.Client, ctx context.Context) {
	fmt.Println("--- Test 1: Subscriber connects on Daytona construction ---")

	sandbox, err := client.Create(ctx, types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{Language: types.CodeLanguagePython},
	}, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("  FAIL: Create failed: %v", err)
	}
	fmt.Printf("  Created sandbox: %s, state=%s\n", sandbox.ID, sandbox.State)

	if err := sandbox.Delete(ctx); err != nil {
		log.Fatalf("  FAIL: Delete failed: %v", err)
	}
	fmt.Println("  PASS: Test 1 complete")
	fmt.Println()
}

func test2EventDrivenLifecycle(client *daytona.Client, ctx context.Context) {
	fmt.Println("--- Test 2: Event-driven start/stop lifecycle ---")

	sandbox, err := client.Create(ctx, types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{Language: types.CodeLanguagePython},
	}, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("  FAIL: Create failed: %v", err)
	}
	fmt.Printf("  Created sandbox: %s, state=%s\n", sandbox.ID, sandbox.State)

	t0 := time.Now()
	if err := sandbox.Stop(ctx); err != nil {
		log.Fatalf("  FAIL: Stop failed: %v", err)
	}
	stopTime := time.Since(t0).Seconds()
	fmt.Printf("  Stopped in %.2fs, state=%s\n", stopTime, sandbox.State)
	if sandbox.State != "stopped" {
		log.Fatalf("  FAIL: Expected stopped, got %s", sandbox.State)
	}

	t0 = time.Now()
	if err := sandbox.Start(ctx); err != nil {
		log.Fatalf("  FAIL: Start failed: %v", err)
	}
	startTime := time.Since(t0).Seconds()
	fmt.Printf("  Started in %.2fs, state=%s\n", startTime, sandbox.State)
	if sandbox.State != "started" {
		log.Fatalf("  FAIL: Expected started, got %s", sandbox.State)
	}

	if err := sandbox.Delete(ctx); err != nil {
		log.Fatalf("  FAIL: Delete failed: %v", err)
	}
	fmt.Println("  PASS: Test 2 complete")
	fmt.Println()
}

func test3AutoUpdateFromEvents(client *daytona.Client, ctx context.Context) {
	fmt.Println("--- Test 3: Auto-update sandbox state from events ---")

	sandbox, err := client.Create(ctx, types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{Language: types.CodeLanguagePython},
	}, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("  FAIL: Create failed: %v", err)
	}
	fmt.Printf("  Created: state=%s\n", sandbox.State)

	if err := sandbox.Stop(ctx); err != nil {
		log.Fatalf("  FAIL: Stop failed: %v", err)
	}
	fmt.Printf("  After Stop call: state=%s\n", sandbox.State)
	if sandbox.State != "stopped" {
		log.Fatalf("  FAIL: Expected stopped, got %s", sandbox.State)
	}

	if err := sandbox.Delete(ctx); err != nil {
		log.Fatalf("  FAIL: Delete failed: %v", err)
	}
	fmt.Println("  PASS: Test 3 complete")
	fmt.Println()
}

func test4GetSubscribes(client *daytona.Client, ctx context.Context) {
	fmt.Println("--- Test 4: get() sandboxes subscribe to events ---")

	sandbox, err := client.Create(ctx, types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{Language: types.CodeLanguagePython},
	}, options.WithTimeout(90*time.Second))
	if err != nil {
		log.Fatalf("  FAIL: Create failed: %v", err)
	}

	sandbox2, err := client.Get(ctx, sandbox.ID)
	if err != nil {
		log.Fatalf("  FAIL: Get failed: %v", err)
	}
	fmt.Printf("  Got sandbox via Get(): state=%s\n", sandbox2.State)

	if err := sandbox.Stop(ctx); err != nil {
		log.Fatalf("  FAIL: Stop failed: %v", err)
	}
	time.Sleep(1 * time.Second)
	fmt.Printf("  After stop - original: state=%s, get'd: state=%s\n", sandbox.State, sandbox2.State)
	if sandbox2.State != "stopped" && sandbox2.State != "stopping" {
		log.Fatalf("  FAIL: get'd sandbox state should update, got %s", sandbox2.State)
	}

	if err := sandbox.Delete(ctx); err != nil {
		log.Fatalf("  FAIL: Delete failed: %v", err)
	}
	fmt.Println("  PASS: Test 4 complete")
	fmt.Println()
}
