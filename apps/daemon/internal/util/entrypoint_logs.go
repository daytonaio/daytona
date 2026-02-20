// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func ReadEntrypointLogs(entrypointLogFilePath string) error {
	if entrypointLogFilePath == "" {
		return errors.New("entrypoint log file path is not configured")
	}

	logFile, err := os.Open(entrypointLogFilePath)
	if err != nil {
		return fmt.Errorf("failed to open entrypoint log file at %s: %w", entrypointLogFilePath, err)
	}
	defer logFile.Close()

	_, err = io.Copy(os.Stdout, logFile)
	if err != nil {
		return fmt.Errorf("failed to read entrypoint log file: %w", err)
	}

	return nil
}
