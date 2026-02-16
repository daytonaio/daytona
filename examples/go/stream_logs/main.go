// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

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

	log.Printf("Created sandbox: %s (ID: %s)\n", sandbox.Name, sandbox.ID)
	defer func() {
		log.Println("\nCleaning up...")
		if err := sandbox.Delete(ctx); err != nil {
			log.Printf("Failed to delete sandbox: %v", err)
		} else {
			log.Println("Sandbox deleted")
		}
	}()

	// Create a session for running async commands
	sessionID := "stream-logs-session"
	err = sandbox.Process.CreateSession(ctx, sessionID)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	log.Printf("Created session: %s\n", sessionID)
	defer func() {
		if err := sandbox.Process.DeleteSession(ctx, sessionID); err != nil {
			log.Printf("Failed to delete session: %v", err)
		}
	}()

	// Execute a command that produces output over time (async)
	log.Println("\n=== Streaming Logs Example ===")
	log.Println("Starting a long-running command asynchronously...")

	// This command outputs to both stdout and stderr over 5 seconds
	cmd := `for i in 1 2 3 4 5; do
		echo "stdout: iteration $i"
		echo "stderr: error message $i" >&2
		sleep 1
	done
	echo "stdout: done!"`

	cmdResult, err := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, cmd, true, false)
	if err != nil {
		log.Fatalf("Failed to execute async command: %v", err)
	}

	cmdID, _ := cmdResult["id"].(string)
	log.Printf("Command started with ID: %s\n", cmdID)

	// Stream the logs as they come in
	log.Println("\nStreaming logs (stdout in green, stderr in red):")
	log.Println("---")

	// Create buffered channels for stdout and stderr
	stdout := make(chan string, 100)
	stderr := make(chan string, 100)

	// Start streaming in a goroutine
	streamErr := make(chan error, 1)
	go func() {
		err := sandbox.Process.GetSessionCommandLogsStream(ctx, sessionID, cmdID, stdout, stderr)
		streamErr <- err
	}()

	// Read from channels until both are closed
	// Buffer partial lines to avoid jumbled output when stdout/stderr interleave
	stdoutOpen, stderrOpen := true, true
	var stdoutBuf, stderrBuf string

	printLines := func(buf *string, prefix, color string, w *os.File) {
		for {
			idx := strings.Index(*buf, "\n")
			if idx == -1 {
				break
			}
			line := (*buf)[:idx+1]
			*buf = (*buf)[idx+1:]
			fmt.Fprintf(w, "%s[%s] %s\033[0m", color, prefix, line)
		}
	}

	for stdoutOpen || stderrOpen {
		select {
		case chunk, ok := <-stdout:
			if !ok {
				stdoutOpen = false
				if stdoutBuf != "" {
					fmt.Fprintf(os.Stdout, "\033[32m[STDOUT] %s\033[0m\n", stdoutBuf)
				}
			} else {
				stdoutBuf += chunk
				printLines(&stdoutBuf, "STDOUT", "\033[32m", os.Stdout)
			}
		case chunk, ok := <-stderr:
			if !ok {
				stderrOpen = false
				if stderrBuf != "" {
					fmt.Fprintf(os.Stderr, "\033[31m[STDERR] %s\033[0m\n", stderrBuf)
				}
			} else {
				stderrBuf += chunk
				printLines(&stderrBuf, "STDERR", "\033[31m", os.Stderr)
			}
		}
	}

	log.Println("---")

	// Check for streaming errors
	if err := <-streamErr; err != nil {
		log.Printf("Stream ended with error: %v", err)
	} else {
		log.Println("Stream completed successfully!")
	}

	// Verify the command completed
	cmdStatus, err := sandbox.Process.GetSessionCommand(ctx, sessionID, cmdID)
	if err != nil {
		log.Fatalf("Failed to get command status: %v", err)
	}
	if exitCode, ok := cmdStatus["exitCode"]; ok {
		log.Printf("Command exit code: %v\n", exitCode)
	}

	log.Println("\nAll streaming logs examples completed successfully!")
}
