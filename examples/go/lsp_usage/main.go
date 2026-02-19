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
	// Create Daytona client
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Create a sandbox with Python
	log.Println("Creating sandbox...")
	sandbox, err := client.Create(ctx, &types.ImageParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language: types.CodeLanguagePython,
		},
		Image: "python:3.11",
		Resources: &types.Resources{
			CPU:    1,
			Memory: 1,
		},
	}, options.WithTimeout(120*time.Second))

	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	defer func() {
		log.Println("\nCleaning up sandbox...")
		_ = sandbox.Delete(ctx)
	}()

	log.Printf("Sandbox created: %s (state: %s)\n", sandbox.Name, sandbox.State)

	// Create a Python file to work with
	log.Println("\nCreating Python file...")
	pythonCode := `def greet(name):
    """Greet someone by name."""
    return f"Hello, {name}!"

def calculate_sum(a, b):
    """Calculate the sum of two numbers."""
    return a + b

# Main code
if __name__ == "__main__":
    print(greet("World"))
    print(calculate_sum(5, 3))
`

	workDir, err := sandbox.GetWorkingDir(ctx)
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	if err := sandbox.FileSystem.UploadFile(ctx, []byte(pythonCode), workDir+"/main.py"); err != nil {
		log.Fatalf("Failed to create Python file: %v", err)
	}

	// Create LSP server for Python
	log.Println("\nCreating LSP server for Python...")
	lsp := daytona.NewLspServerService(sandbox.ToolboxClient, types.LspLanguagePython, workDir, client.Otel)

	// Start the LSP server
	log.Println("Starting LSP server...")
	if err := lsp.Start(ctx); err != nil {
		log.Fatalf("Failed to start LSP server: %v", err)
	}
	defer func() {
		log.Println("Stopping LSP server...")
		_ = lsp.Stop(ctx)
	}()

	// Notify LSP about the opened file
	log.Println("Notifying LSP about opened file...")
	if err := lsp.DidOpen(ctx, "main.py"); err != nil {
		log.Fatalf("Failed to notify LSP: %v", err)
	}

	// Get document symbols
	log.Println("\nGetting document symbols...")
	symbols, err := lsp.DocumentSymbols(ctx, "main.py")
	if err != nil {
		log.Fatalf("Failed to get document symbols: %v", err)
	}

	log.Printf("\nFound %d symbols in main.py\n", len(symbols))
	if len(symbols) > 0 {
		log.Println("Document symbols retrieved successfully!")
	}

	// Get completions at a specific position
	log.Println("\nGetting code completions at line 10, character 10...")
	completions, err := lsp.Completions(ctx, "main.py", types.Position{Line: 10, Character: 10})
	if err != nil {
		log.Fatalf("Failed to get completions: %v", err)
	}

	if completions != nil {
		log.Println("Code completions retrieved successfully!")
	} else {
		log.Println("No completions available")
	}

	// TODO: why is search for sandbox symbols not working?
	// Search for symbols in the workspace
	// Note: workspace-symbols endpoint may not be available in all environments
	log.Println("\nSearching for 'greet' in workspace...")
	workspaceSymbols, err := lsp.SandboxSymbols(ctx, "greet")
	if err != nil {
		log.Printf("Workspace symbols search not available: %v\n", err)
	} else {
		log.Printf("Found %d workspace symbols matching 'greet'\n", len(workspaceSymbols))
	}

	// Close the file
	log.Println("\nClosing file...")
	if err := lsp.DidClose(ctx, "main.py"); err != nil {
		log.Fatalf("Failed to close file: %v", err)
	}

	log.Println("\n✅ LSP server demo completed successfully!")
}
