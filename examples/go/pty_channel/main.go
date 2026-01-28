// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
	// Create Daytona client
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

	sandbox, err := client.Create(ctx, params)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	defer func() {
		_ = sandbox.Delete(ctx)
	}()

	log.Printf("✓ Created sandbox: %s\n", sandbox.ID)

	// Create a PTY session
	handle, err := sandbox.Process.CreatePty(ctx, "demo-session")
	if err != nil {
		log.Fatalf("Failed to create PTY: %v", err)
	}
	defer func() {
		_ = handle.Disconnect()
	}()

	// Wait for connection
	if err := handle.WaitForConnection(ctx); err != nil {
		log.Fatalf("Failed to wait for connection: %v", err)
	}

	log.Println("✓ Connected to PTY")

	// Read output from the channel
	go func() {
		for data := range handle.DataChan() {
			fmt.Print(string(data))
		}
	}()

	// Send some commands
	_ = handle.SendInput([]byte("echo 'Hello from PTY!'\n"))
	_ = handle.SendInput([]byte("pwd\n"))
	_ = handle.SendInput([]byte("ls -la\n"))
	_ = handle.SendInput([]byte("exit\n"))

	// Wait for the PTY to exit
	result, err := handle.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to wait for PTY: %v", err)
	}

	log.Printf("\n✓ PTY exited with code: %d\n", *result.ExitCode)
}
