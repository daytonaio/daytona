// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/base64"
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

	// Create a sandbox with desktop environment support
	log.Println("Creating sandbox with desktop environment...")
	params := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language: types.CodeLanguagePython,
		},
	}

	sandbox, err := client.Create(ctx, params, options.WithTimeout(90*time.Second))
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

	// Start the desktop environment
	log.Println("\nStarting desktop environment...")
	if err := sandbox.ComputerUse.Start(ctx); err != nil {
		log.Fatalf("Failed to start desktop environment: %v", err)
	}
	log.Println("✓ Desktop environment started")

	// Wait for desktop to be ready
	time.Sleep(3 * time.Second)

	// Check desktop status
	status, err := sandbox.ComputerUse.GetStatus(ctx)
	if err != nil {
		log.Fatalf("Failed to get desktop status: %v", err)
	}
	log.Printf("Desktop status: %v\n", status)

	// Example 1: Display Information
	log.Println("\n=== Display Information ===")
	displayInfo, err := sandbox.ComputerUse.Display().GetInfo(ctx)
	if err != nil {
		log.Fatalf("Failed to get display info: %v", err)
	}
	log.Printf("Display info: %v\n", displayInfo)

	// Example 2: Mouse Operations
	log.Println("\n=== Mouse Operations ===")

	// Get current mouse position
	mousePos, err := sandbox.ComputerUse.Mouse().GetPosition(ctx)
	if err != nil {
		log.Fatalf("Failed to get mouse position: %v", err)
	}
	log.Printf("Current mouse position: %v\n", mousePos)

	// Move mouse to a specific position
	newPos, err := sandbox.ComputerUse.Mouse().Move(ctx, 100, 100)
	if err != nil {
		log.Fatalf("Failed to move mouse: %v", err)
	}
	log.Printf("✓ Moved mouse to: %v\n", newPos)

	// Click at current position
	clickPos, err := sandbox.ComputerUse.Mouse().Click(ctx, 100, 100, nil, nil)
	if err != nil {
		log.Fatalf("Failed to click mouse: %v", err)
	}
	log.Printf("✓ Clicked at: %v\n", clickPos)

	// Click with specific button (left, right, middle)
	leftButton := "left"
	doubleClick := true
	_, err = sandbox.ComputerUse.Mouse().Click(ctx, 150, 150, &leftButton, &doubleClick)
	if err != nil {
		log.Fatalf("Failed to double click: %v", err)
	}
	log.Println("✓ Double-clicked with left button")

	// Drag operation
	dragResult, err := sandbox.ComputerUse.Mouse().Drag(ctx, 100, 100, 200, 200, &leftButton)
	if err != nil {
		log.Fatalf("Failed to drag: %v", err)
	}
	log.Printf("✓ Dragged from (100,100) to (200,200), final position: %v\n", dragResult)

	// Scroll operation
	// TOFIX: timing out?
	// amount := 3
	// scrollSuccess, err := sandbox.ComputerUse.Mouse().Scroll(ctx, 300, 300, "down", &amount)
	// if err != nil {
	// 	log.Fatalf("Failed to scroll: %v", err)
	// }
	// log.Printf("✓ Scrolled down, success: %v\n", scrollSuccess)

	// Example 3: Keyboard Operations
	log.Println("\n=== Keyboard Operations ===")

	// Type text
	if err := sandbox.ComputerUse.Keyboard().Type(ctx, "Hello, Daytona!", nil); err != nil {
		log.Fatalf("Failed to type text: %v", err)
	}
	log.Println("✓ Typed: 'Hello, Daytona!'")

	// Type with delay between characters
	delay := 50 // milliseconds
	if err := sandbox.ComputerUse.Keyboard().Type(ctx, "Slow typing...", &delay); err != nil {
		log.Fatalf("Failed to type with delay: %v", err)
	}
	log.Println("✓ Typed with delay: 'Slow typing...'")

	// Press a key with modifiers
	modifiers := []string{"ctrl"}
	if err := sandbox.ComputerUse.Keyboard().Press(ctx, "c", modifiers); err != nil {
		log.Fatalf("Failed to press key: %v", err)
	}
	log.Println("✓ Pressed: Ctrl+C")

	// Press multiple modifiers
	modifiers = []string{"ctrl", "shift"}
	if err := sandbox.ComputerUse.Keyboard().Press(ctx, "t", modifiers); err != nil {
		log.Fatalf("Failed to press key combo: %v", err)
	}
	log.Println("✓ Pressed: Ctrl+Shift+T")

	// Execute hotkey
	if err := sandbox.ComputerUse.Keyboard().Hotkey(ctx, "alt+tab"); err != nil {
		log.Fatalf("Failed to press hotkey: %v", err)
	}
	log.Println("✓ Pressed hotkey: Alt+Tab")

	// Example 4: Screenshots
	log.Println("\n=== Screenshot Operations ===")

	// Take full screen screenshot
	showCursor := true
	screenshot, err := sandbox.ComputerUse.Screenshot().TakeFullScreen(ctx, &showCursor)
	if err != nil {
		log.Fatalf("Failed to take screenshot: %v", err)
	}
	log.Printf("✓ Full screen screenshot taken (cursor visible)\n")

	// Save screenshot to file
	if screenshot.Image != "" {
		screenshotData, err := base64.StdEncoding.DecodeString(screenshot.Image)
		if err != nil {
			log.Printf("Warning: Failed to decode screenshot: %v\n", err)
		} else {
			filename := "fullscreen_screenshot.png"
			if err := os.WriteFile(filename, screenshotData, 0644); err != nil {
				log.Printf("Warning: Failed to save screenshot: %v\n", err)
			} else {
				log.Printf("✓ Saved screenshot to: %s\n", filename)
			}
		}
	}

	// TOFIX: needs backend fix for this to work
	// // Take regional screenshot
	// region := types.ScreenshotRegion{
	// 	X:      100,
	// 	Y:      100,
	// 	Width:  400,
	// 	Height: 300,
	// }
	// hideCursor := false
	// regionScreenshot, err := sandbox.ComputerUse.Screenshot().TakeRegion(ctx, region, &hideCursor)
	// if err != nil {
	// 	log.Fatalf("Failed to take region screenshot: %v", err)
	// }
	// log.Printf("✓ Regional screenshot taken (region: %dx%d at %d,%d)\n",
	// 	region.Width, region.Height, region.X, region.Y)

	// // Save region screenshot
	// if regionScreenshot.Image != "" {
	// 	screenshotData, err := base64.StdEncoding.DecodeString(regionScreenshot.Image)
	// 	if err != nil {
	// 		log.Printf("Warning: Failed to decode region screenshot: %v\n", err)
	// 	} else {
	// 		filename := "region_screenshot.png"
	// 		if err := os.WriteFile(filename, screenshotData, 0644); err != nil {
	// 			log.Printf("Warning: Failed to save region screenshot: %v\n", err)
	// 		} else {
	// 			log.Printf("✓ Saved region screenshot to: %s\n", filename)
	// 		}
	// 	}
	// }

	// Example 5: Window Management
	log.Println("\n=== Window Management ===")
	windows, err := sandbox.ComputerUse.Display().GetWindows(ctx)
	if err != nil {
		log.Fatalf("Failed to get windows: %v", err)
	}
	log.Printf("Open windows: %v\n", windows)

	// Stop the desktop environment
	log.Println("\nStopping desktop environment...")
	if err := sandbox.ComputerUse.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop desktop environment: %v", err)
	}
	log.Println("✓ Desktop environment stopped")
	log.Println("\n✓ All computer use operations completed successfully!")
}
