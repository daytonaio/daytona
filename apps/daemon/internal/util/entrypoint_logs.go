// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"fmt"
	"io"
	"os"
)

func ReadEntrypointLogs(entrypointLogFilePath string) {
	if entrypointLogFilePath == "" {
		fmt.Fprintln(os.Stderr, "Error: Entrypoint log file path is not configured")
		os.Exit(1)
	}

	logFile, err := os.Open(entrypointLogFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open entrypoint log file at %s: %v\n", entrypointLogFilePath, err)
		os.Exit(1)
	}
	defer logFile.Close()

	_, err = io.Copy(os.Stdout, logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read entrypoint log file: %v\n", err)
		os.Exit(1)
	}
}
