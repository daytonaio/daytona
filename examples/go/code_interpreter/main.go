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

	// Create a sandbox with Python
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

	log.Printf("✓ Created sandbox: %s (ID: %s)\n\n", sandbox.Name, sandbox.ID)

	// Get the code interpreter service for this sandbox
	interpreter := sandbox.CodeInterpreter

	// Example 1: Simple code execution
	log.Println("=== Example 1: Simple Python execution ===")
	channels, err := interpreter.RunCode(
		ctx,
		"print('Hello from Daytona!')\nprint('Python version:', __import__('sys').version)",
	)
	if err != nil {
		log.Fatalf("Failed to run code: %v", err)
	}

	// Wait for execution to complete
	result := <-channels.Done
	log.Printf("Output:\n%s\n", result.Stdout)
	if result.Stderr != "" {
		log.Printf("Stderr:\n%s\n", result.Stderr)
	}
	if result.Error != nil {
		log.Printf("Error: %s - %s\n", result.Error.Name, result.Error.Value)
	}

	// Example 2: Execution with real-time streaming
	log.Println("\n=== Example 2: Execution with real-time streaming ===")

	channels, err = interpreter.RunCode(
		ctx,
		`import time
for i in range(5):
    print(f"Processing step {i+1}...")
    time.sleep(0.5)
print("Done!")`,
	)
	if err != nil {
		log.Fatalf("Failed to run code: %v", err)
	}

	// Start goroutines to read from channels in real-time
	go func() {
		for msg := range channels.Stdout {
			log.Printf("[STDOUT] %s", msg.Text)
		}
	}()
	go func() {
		for msg := range channels.Stderr {
			log.Printf("[STDERR] %s", msg.Text)
		}
	}()
	go func() {
		for execErr := range channels.Errors {
			log.Printf("[ERROR] %s: %s\n", execErr.Name, execErr.Value)
			if execErr.Traceback != nil {
				log.Printf("Traceback:\n%s\n", *execErr.Traceback)
			}
		}
	}()

	// Wait for execution to complete
	<-channels.Done
	log.Printf("Execution completed\n")

	// Example 3: Code execution with environment variables
	log.Println("\n=== Example 3: Using environment variables ===")
	env := map[string]string{
		"MY_VAR":    "Hello",
		"MY_NUMBER": "42",
	}
	channels, err = interpreter.RunCode(
		ctx,
		`import os
print(f"MY_VAR: {os.environ.get('MY_VAR')}")
print(f"MY_NUMBER: {os.environ.get('MY_NUMBER')}")`,
		options.WithEnv(env),
	)
	if err != nil {
		log.Fatalf("Failed to run code: %v", err)
	}
	result = <-channels.Done
	log.Printf("Output:\n%s\n", result.Stdout)

	// Example 4: Error handling with channels
	log.Println("\n=== Example 4: Error handling with channels ===")

	channels, err = interpreter.RunCode(
		ctx,
		`# This will cause a runtime error
x = 1 / 0`,
	)
	if err != nil {
		log.Fatalf("Failed to run code: %v", err)
	}

	// Read from error channel in real-time
	go func() {
		for execErr := range channels.Errors {
			log.Printf("Caught error: %s\n", execErr.Name)
			log.Printf("Message: %s\n", execErr.Value)
			if execErr.Traceback != nil {
				log.Printf("Traceback:\n%s\n", *execErr.Traceback)
			}
		}
	}()

	result = <-channels.Done
	if result.Error != nil {
		log.Printf("Execution completed with error: %s\n", result.Error.Name)
	}

	// Example 5: Using custom timeout
	log.Println("\n=== Example 5: Execution with timeout ===")

	channels, err = interpreter.RunCode(
		ctx,
		`import time
print("Starting long operation...")
time.sleep(10)  # This will timeout after 2 seconds
print("This won't be printed")`,
		options.WithInterpreterTimeout(2*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to run code: %v", err)
	}

	// Read stdout in real-time
	go func() {
		for msg := range channels.Stdout {
			log.Printf("[STDOUT] %s", msg.Text)
		}
	}()

	result = <-channels.Done
	if result.Error != nil {
		log.Printf("Expected timeout error: %s - %s\n", result.Error.Name, result.Error.Value)
	}

	// Example 6: Working with data and computations
	log.Println("\n=== Example 6: Data processing ===")
	channels, err = interpreter.RunCode(
		ctx,
		`import json

# Simulate data processing
data = [1, 2, 3, 4, 5]
squared = [x**2 for x in data]
print(f"Original: {data}")
print(f"Squared: {squared}")
print(f"Sum: {sum(squared)}")

# Output as JSON
result = {
    "original": data,
    "squared": squared,
    "sum": sum(squared)
}
print(json.dumps(result, indent=2))`,
	)
	if err != nil {
		log.Fatalf("Failed to run code: %v", err)
	}
	result = <-channels.Done
	log.Printf("Output:\n%s\n", result.Stdout)

	// Example 7: Creating and using a custom context
	log.Println("\n=== Example 7: Using custom interpreter context ===")
	cwd := "/tmp"
	contextInfo, err := interpreter.CreateContext(ctx, &cwd)
	if err != nil {
		log.Fatalf("Failed to create context: %v", err)
	}
	log.Printf("Created context: %v\n", contextInfo["id"])

	// Use the custom context
	contextID := contextInfo["id"].(string)
	channels, err = interpreter.RunCode(
		ctx,
		`import os
print(f"Current directory: {os.getcwd()}")
print(f"Files: {os.listdir('.')[:5]}")  # Show first 5 files`,
		options.WithCustomContext(contextID),
	)
	if err != nil {
		log.Fatalf("Failed to run code in custom context: %v", err)
	}
	result = <-channels.Done
	log.Printf("Output:\n%s\n", result.Stdout)

	// Delete the sandbox
	log.Println("\n=== Cleaning up ===")
	if err := sandbox.Delete(ctx); err != nil {
		log.Fatalf("Failed to delete sandbox: %v", err)
	}
	log.Println("✓ Sandbox deleted")

	log.Println("\n✓ All code interpreter examples completed successfully!")
}
