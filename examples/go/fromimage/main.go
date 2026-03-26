// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"os"
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

	// Example 1: Create a sandbox from a simple base image string
	log.Println("=== Example 1: Creating sandbox from base image string ===")
	params1 := types.ImageParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Name: "simple-image-sandbox",
			EnvVars: map[string]string{
				"ENV": "production",
			},
		},
		Image: "python:3.12-slim",
		Resources: &types.Resources{
			CPU:    1,
			Memory: 1,
		},
	}

	logChan := make(chan string, 100)
	go func() {
		for logLine := range logChan {
			log.Printf("[BUILD] %s\n", logLine)
		}
	}()

	sandbox1, err := client.Create(ctx, params1, options.WithTimeout(120*time.Second), options.WithLogChannel(logChan))
	if err != nil {
		log.Fatalf("Failed to create sandbox from image string: %v", err)
	}

	log.Printf("✓ Created sandbox: %s (ID: %s)\n\n", sandbox1.Name, sandbox1.ID)

	// Clean up first sandbox
	defer func() {
		log.Println("Cleaning up first sandbox...")
		if err := sandbox1.Delete(ctx); err != nil {
			log.Printf("Failed to delete sandbox: %v", err)
		} else {
			log.Println("✓ First sandbox deleted")
		}
	}()

	// Example 2: Create a custom image with Python packages
	log.Println("=== Example 2: Creating custom image with Python packages ===")
	image2 := daytona.Base("python:3.12-slim-bookworm").
		PipInstall([]string{"numpy", "pandas", "matplotlib"}, options.WithFindLinks("https://pypi.org/simple")).
		Env("APP_ENV", "development").
		Workdir("/app")

	params2 := types.ImageParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Name: "custom-python-sandbox",
		},
		Image: image2,
		Resources: &types.Resources{
			CPU:    1,
			Memory: 1,
		},
	}

	logChan2 := make(chan string, 100)
	go func() {
		for logLine := range logChan2 {
			log.Printf("[BUILD] %s\n", logLine)
		}
	}()
	sandbox2, err := client.Create(ctx, params2, options.WithTimeout(120*time.Second), options.WithLogChannel(logChan2))
	if err != nil {
		log.Fatalf("Failed to create sandbox with custom image: %v", err)
	}

	log.Printf("✓ Created sandbox: %s (ID: %s)\n\n", sandbox2.Name, sandbox2.ID)

	defer func() {
		log.Println("Cleaning up second sandbox...")
		if err := sandbox2.Delete(ctx); err != nil {
			log.Printf("Failed to delete sandbox: %v", err)
		} else {
			log.Println("✓ Second sandbox deleted")
		}
	}()

	// Verify packages are installed
	log.Println("Verifying Python packages are installed...")
	result, err := sandbox2.Process.ExecuteCommand(ctx, "python3 -c 'import numpy, pandas, matplotlib; print(\"All packages imported successfully!\")'")
	if err != nil {
		log.Printf("Failed to verify packages: %v", err)
	} else {
		log.Printf("✓ %s\n\n", result.Result)
	}

	// Example 3: Create an image with local files (if available)
	log.Println("=== Example 3: Creating image with local files ===")

	// Create temporary test files for demonstration
	tmpDir := "/tmp/daytona-example"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		log.Printf("Warning: Could not create temp directory: %v", err)
	} else {
		// Create a sample config file
		configPath := tmpDir + "/config.json"
		configContent := []byte(`{
  "app_name": "Daytona Example",
  "version": "1.0.0",
  "features": {
    "logging": true,
    "metrics": true
  }
}`)
		if err := os.WriteFile(configPath, configContent, 0644); err != nil {
			log.Printf("Warning: Could not create config file: %v", err)
		}

		// Create a sample Python script
		scriptPath := tmpDir + "/app.py"
		scriptContent := []byte(`#!/usr/bin/env python3
import json
import os

def main():
    config_path = '/app/config.json'
    if os.path.exists(config_path):
        with open(config_path, 'r') as f:
            config = json.load(f)
        print(f"App: {config['app_name']}")
        print(f"Version: {config['version']}")
        print("Configuration loaded successfully!")
    else:
        print("Config file not found!")

if __name__ == '__main__':
    main()
`)
		if err := os.WriteFile(scriptPath, scriptContent, 0755); err != nil {
			log.Printf("Warning: Could not create script file: %v", err)
		}

		// Create image with local files
		image3 := daytona.Base("python:3.12-slim-bookworm").
			AddLocalFile(configPath, "/app/config.json").
			AddLocalFile(scriptPath, "/app/app.py").
			Workdir("/app").
			Run("chmod +x /app/app.py")

		params3 := types.ImageParams{
			SandboxBaseParams: types.SandboxBaseParams{
				Name: "sandbox-with-files",
			},
			Image: image3,
		}

		logChan3 := make(chan string, 100)
		go func() {
			for logLine := range logChan3 {
				log.Printf("[BUILD] %s\n", logLine)
			}
		}()
		sandbox3, err := client.Create(ctx, params3, options.WithTimeout(120*time.Second), options.WithLogChannel(logChan3))
		if err != nil {
			log.Fatalf("Failed to create sandbox with local files: %v", err)
		}

		log.Printf("✓ Created sandbox: %s (ID: %s)\n\n", sandbox3.Name, sandbox3.ID)

		defer func() {
			log.Println("Cleaning up third sandbox...")
			if err := sandbox3.Delete(ctx); err != nil {
				log.Printf("Failed to delete sandbox: %v", err)
			} else {
				log.Println("✓ Third sandbox deleted")
			}
			// Clean up temp files
			os.RemoveAll(tmpDir)
		}()

		// Verify files were copied
		log.Println("Verifying local files were copied...")
		result, err := sandbox3.Process.ExecuteCommand(ctx, "python3 /app/app.py")
		if err != nil {
			log.Printf("Failed to run app: %v", err)
		} else {
			log.Printf("✓ App output:\n%s\n\n", result.Result)
		}

		// List files in /app
		files, err := sandbox3.FileSystem.ListFiles(ctx, "/app")
		if err != nil {
			log.Printf("Failed to list files: %v", err)
		} else {
			log.Println("Files in /app:")
			for _, file := range files {
				log.Printf("  - %s (%d bytes)\n", file.Name, file.Size)
			}
		}
	}

	// Example 4: Using DebianSlim helper
	log.Println("\n=== Example 4: Using DebianSlim helper ===")
	pythonVersion := "3.11"
	image4 := daytona.DebianSlim(&pythonVersion).
		PipInstall([]string{"requests", "flask"}, options.WithFindLinks("https://pypi.org/simple")).
		Expose([]int{5000}).
		Env("FLASK_APP", "app.py").
		Workdir("/app")

	params4 := types.ImageParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Name: "debian-slim-sandbox",
		},
		Image: image4,
	}

	logChan4 := make(chan string, 100)
	go func() {
		for logLine := range logChan4 {
			log.Printf("[BUILD] %s\n", logLine)
		}
	}()
	sandbox4, err := client.Create(ctx, params4, options.WithTimeout(120*time.Second), options.WithLogChannel(logChan4))
	if err != nil {
		log.Fatalf("Failed to create DebianSlim sandbox: %v", err)
	}

	log.Printf("✓ Created sandbox: %s (ID: %s)\n", sandbox4.Name, sandbox4.ID)

	defer func() {
		log.Println("Cleaning up fourth sandbox...")
		if err := sandbox4.Delete(ctx); err != nil {
			log.Printf("Failed to delete sandbox: %v", err)
		} else {
			log.Println("✓ Fourth sandbox deleted")
		}
	}()

	// Verify Python version
	result, err = sandbox4.Process.ExecuteCommand(ctx, "python3 --version")
	if err != nil {
		log.Printf("Failed to get Python version: %v", err)
	} else {
		log.Printf("Python version: %s\n", result.Result)
	}

	// Verify packages
	result, err = sandbox4.Process.ExecuteCommand(ctx, "bash -c \"pip list | grep -E '(requests|flask)'\"")
	if err != nil {
		log.Printf("Failed to list packages: %v", err)
	} else {
		log.Printf("Installed packages:\n%s\n", result.Result)
	}

	log.Println("\n✓ All image builder examples completed successfully!")
}
