// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

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
	// Create Daytona client (uses DAYTONA_API_KEY from environment)
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Create a volume with a unique name
	volumeName := fmt.Sprintf("data-volume-%d", time.Now().Unix())
	log.Printf("Creating volume: %s\n", volumeName)
	volume, err := client.Volume.Create(ctx, volumeName)
	if err != nil {
		log.Fatalf("Failed to create volume: %v", err)
	}
	log.Printf("✓ Created volume: %s (ID: %s, State: %s)\n", volume.Name, volume.ID, volume.State)

	// Wait for volume to be ready
	log.Println("Waiting for volume to become ready...")
	volume, err = client.Volume.WaitForReady(ctx, volume, 60*time.Second)
	if err != nil {
		log.Fatalf("Failed waiting for volume: %v", err)
	}
	log.Printf("✓ Volume is now ready (State: %s)\n", volume.State)

	// Create a sandbox with the volume mounted
	log.Printf("Creating sandbox with volume ID: %s mounted at /data\n", volume.ID)
	params := types.ImageParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Name:     "sandbox-with-volume",
			Language: types.CodeLanguagePython,
			Volumes: []types.VolumeMount{
				{
					VolumeID:  volume.ID,
					MountPath: "/data",
					Subpath:   nil, // Mount entire volume
				},
			},
		},
		Image: "python:3.12-slim",
	}

	logChan := make(chan string, 100)
	go func() {
		for logLine := range logChan {
			log.Printf("[BUILD] %s\n", logLine)
		}
	}()
	sandbox, err := client.Create(ctx, params, options.WithLogChannel(logChan))
	if err != nil {
		log.Printf("Failed to create sandbox with volume: %v", err)
		// Clean up volume and exit
		_ = client.Volume.Delete(ctx, volume)
		log.Fatalf("Exiting due to error")
	}

	log.Printf("✓ Created sandbox: %s (ID: %s)\n", sandbox.Name, sandbox.ID)

	// Verify volume is mounted by listing the directory
	result, err := sandbox.Process.ExecuteCommand(ctx, "ls -la /data")
	if err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}
	log.Printf("Volume mounted at /data:\n%s\n", result.Result)

	// Write a file to the volume
	_, err = sandbox.Process.ExecuteCommand(ctx, "echo 'Hello from volume!' > /data/test.txt")
	if err != nil {
		log.Fatalf("Failed to write to volume: %v", err)
	}
	log.Printf("✓ Wrote file to volume\n")

	// Read the file back
	result, err = sandbox.Process.ExecuteCommand(ctx, "cat /data/test.txt")
	if err != nil {
		log.Fatalf("Failed to read from volume: %v", err)
	}
	log.Printf("File content: %s\n", result.Result)

	// Clean up
	log.Println("Cleaning up...")
	if err := sandbox.Delete(ctx); err != nil {
		log.Printf("Failed to delete sandbox: %v", err)
	} else {
		log.Println("✓ Deleted sandbox")
	}

	// Delete volume (should still be in active state)
	if err := client.Volume.Delete(ctx, volume); err != nil {
		log.Printf("Failed to delete volume: %v", err)
	} else {
		log.Println("✓ Deleted volume")
	}

	log.Println("✓ Done")
}
