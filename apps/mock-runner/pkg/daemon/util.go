// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteStaticBinary extracts the embedded binary to a temp file and returns its path
func WriteStaticBinary(name string) (string, error) {
	data, err := staticAssets.ReadFile(fmt.Sprintf("static/%s", name))
	if err != nil {
		return "", fmt.Errorf("failed to read embedded binary %s: %w", name, err)
	}

	tmpDir := os.TempDir()
	destPath := filepath.Join(tmpDir, name)

	err = os.WriteFile(destPath, data, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to write binary to %s: %w", destPath, err)
	}

	return destPath, nil
}
