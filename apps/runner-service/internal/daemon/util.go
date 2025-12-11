/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteStaticBinary extracts an embedded binary to the filesystem
// Returns the absolute path to the extracted binary
func WriteStaticBinary(name string) (string, error) {
	// Read embedded binary
	binaryData, err := static.ReadFile(fmt.Sprintf("static/%s", name))
	if err != nil {
		return "", fmt.Errorf("failed to read embedded binary %s: %w", name, err)
	}

	// Get working directory
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create temporary binaries directory
	tmpBinariesDir := filepath.Join(pwd, ".tmp", "binaries")
	if err := os.MkdirAll(tmpBinariesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create binaries directory: %w", err)
	}

	// Define binary path
	binaryPath := filepath.Join(tmpBinariesDir, name)

	// Remove existing binary if present
	if _, err := os.Stat(binaryPath); err == nil {
		if err := os.Remove(binaryPath); err != nil {
			return "", fmt.Errorf("failed to remove existing binary: %w", err)
		}
	}

	// Write binary with executable permissions
	if err := os.WriteFile(binaryPath, binaryData, 0755); err != nil {
		return "", fmt.Errorf("failed to write binary: %w", err)
	}

	return binaryPath, nil
}
