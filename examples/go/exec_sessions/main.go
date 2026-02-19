// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
	// Create a new Daytona client
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

	log.Printf("✓ Created sandbox: %s (ID: %s)\n", sandbox.Name, sandbox.ID)
	defer func() {
		log.Println("\nCleaning up...")
		if err := sandbox.Delete(ctx); err != nil {
			log.Printf("Failed to delete sandbox: %v", err)
		} else {
			log.Println("✓ Sandbox deleted")
		}
	}()

	// Example 1: Create an exec session
	log.Println("\n=== Example 1: Creating and using an exec session ===")
	sessionID := "exec-session-1"

	// Create the session
	err = sandbox.Process.CreateSession(ctx, sessionID)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	log.Printf("✓ Created session: %s\n", sessionID)

	// Get session details
	sessionDetails, err := sandbox.Process.GetSession(ctx, sessionID)
	if err != nil {
		log.Fatalf("Failed to get session: %v", err)
	}
	log.Printf("Session details: %+v\n", sessionDetails)

	// Execute a first command in the session
	log.Println("\nExecuting first command in session...")
	cmd1, err := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, "export FOO=BAR", false, false)
	if err != nil {
		log.Fatalf("Failed to execute session command: %v", err)
	}
	log.Printf("✓ Command executed (ID: %s)\n", cmd1["id"])

	// Get the updated session details
	sessionDetails, err = sandbox.Process.GetSession(ctx, sessionID)
	if err != nil {
		log.Fatalf("Failed to get session: %v", err)
	}
	log.Printf("Session updated with command: %v\n", sessionDetails)

	// Execute a second command to verify environment variable
	log.Println("\nExecuting second command in session...")
	cmd2, err := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, "echo $FOO", false, false)
	if err != nil {
		log.Fatalf("Failed to execute second command: %v", err)
	}
	if stdout, ok := cmd2["stdout"].(string); ok {
		log.Printf("FOO=%s\n", stdout)
	}

	// Get logs for the command
	cmdID, _ := cmd2["id"].(string)
	logs, err := sandbox.Process.GetSessionCommandLogs(ctx, sessionID, cmdID)
	if err != nil {
		log.Fatalf("Failed to get command logs: %v", err)
	}
	if logContent, ok := logs["logs"].(string); ok {
		log.Printf("Command logs: %s\n", logContent)
	}

	// Delete the session
	if err := sandbox.Process.DeleteSession(ctx, sessionID); err != nil {
		log.Fatalf("Failed to delete session: %v", err)
	}
	log.Println("✓ Session deleted")

	// Example 2: Session execution with async command
	log.Println("\n=== Example 2: Session with Async Command Execution ===")
	asyncSessionID := "exec-session-async"
	err = sandbox.Process.CreateSession(ctx, asyncSessionID)
	if err != nil {
		log.Fatalf("Failed to create async session: %v", err)
	}

	log.Println("Executing long running command asynchronously...")
	cmd := "counter=1; while (( counter <= 3 )); do echo \"Count: $counter\"; ((counter++)); sleep 1; done"
	cmdResult, err := sandbox.Process.ExecuteSessionCommand(ctx, asyncSessionID, cmd, true, false)
	if err != nil {
		log.Fatalf("Failed to execute async command: %v", err)
	}

	cmdIDAsync, _ := cmdResult["id"].(string)
	log.Printf("Command started with ID: %s\n", cmdIDAsync)

	// Poll for command completion
	log.Println("Waiting for command to complete...")
	time.Sleep(4 * time.Second)

	// Get command status
	cmdStatus, err := sandbox.Process.GetSessionCommand(ctx, asyncSessionID, cmdIDAsync)
	if err != nil {
		log.Fatalf("Failed to get command status: %v", err)
	}
	log.Printf("Command status: %+v\n", cmdStatus)

	// Get logs after completion
	logsAsync, err := sandbox.Process.GetSessionCommandLogs(ctx, asyncSessionID, cmdIDAsync)
	if err != nil {
		log.Fatalf("Failed to get logs: %v", err)
	}
	if logContent, ok := logsAsync["logs"].(string); ok {
		log.Printf("Command logs:\n%s\n", logContent)
	}

	// Delete the async session
	if err := sandbox.Process.DeleteSession(ctx, asyncSessionID); err != nil {
		log.Fatalf("Failed to delete async session: %v", err)
	}
	log.Println("✓ Async session deleted")

	log.Println("\n✓ All session examples completed successfully!")
}
