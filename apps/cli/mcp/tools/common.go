// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"os"
	"path/filepath"
)

var daytonaMCPHeaders map[string]string = map[string]string{
	"X-Daytona-Source": "daytona-mcp",
}

func getSandboxFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".daytona", "sandbox.track"), nil
}

// Helper functions for sandbox tracking
func getActiveSandbox() (string, error) {
	path, err := getSandboxFilePath()
	if err != nil {
		return "", err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func setActiveSandbox(id string) error {
	path, err := getSandboxFilePath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(id), 0644)
}

func clearActiveSandbox() error {
	path, err := getSandboxFilePath()
	if err != nil {
		return err
	}

	return os.Remove(path)
}
