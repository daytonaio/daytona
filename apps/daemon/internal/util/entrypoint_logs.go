// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/daytonaio/common-go/pkg/log"
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

	entrypointLogFile, err := os.Open(entrypointLogFilePath)
	if err != nil {
		return err
	}
	defer entrypointLogFile.Close()

	errChan := make(chan error, 1)
	stdoutChan := make(chan []byte)
	stderrChan := make(chan []byte)
	go log.ReadMultiplexedLog(context.Background(), entrypointLogFile, true, stdoutChan, stderrChan, errChan)
	for {
		select {
		case line := <-stdoutChan:
			_, err := os.Stdout.Write(line)
			if err != nil {
				return fmt.Errorf("failed to write entrypoint log line to stdout: %w", err)
			}
		case line := <-stderrChan:
			_, err := os.Stderr.Write(line)
			if err != nil {
				return fmt.Errorf("failed to write entrypoint log line to stderr: %w", err)
			}
		case err := <-errChan:
			if err != nil {
				return err
			}
		}
	}
}
