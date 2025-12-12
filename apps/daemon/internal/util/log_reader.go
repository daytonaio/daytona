// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"time"
)

func ReadLogWithExitCode(ctx context.Context, logReader io.Reader, follow bool, exitCodeFilePath string, c chan []byte, errChan chan error) {
	reader := bufio.NewReader(logReader)
	consecutiveEOFCount := 0
	maxConsecutiveEOF := 50 // Check exit code after 50 consecutive EOF reads ( 50 * 20ms = 1 second)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			bytes := make([]byte, 1024)
			n, err := reader.Read(bytes)

			if err != nil {
				if err != io.EOF {
					errChan <- err
					return
				} else if !follow {
					errChan <- io.EOF
					return
				}

				// EOF while following - increment counter
				consecutiveEOFCount++

				// Check exit code after maxConsecutiveEOF consecutive EOF reads
				if exitCodeFilePath != "" && consecutiveEOFCount >= maxConsecutiveEOF {
					hasExit := hasExitCode(exitCodeFilePath)
					if hasExit {
						errChan <- io.EOF
						return
					}
					// Reset counter and continue
					consecutiveEOFCount = 0
				}

				// Sleep for a short time to avoid busy-waiting
				time.Sleep(20 * time.Millisecond)
				continue
			}

			// Reset EOF counter on successful read
			if consecutiveEOFCount > 0 {
				consecutiveEOFCount = 0
			}

			if n > 0 {
				// Create a new slice with only the actual read data to avoid sending null bytes
				data := make([]byte, n)
				copy(data, bytes[:n])
				c <- data
			}
		}
	}
}

func hasExitCode(exitCodeFilePath string) bool {
	content, err := os.ReadFile(exitCodeFilePath)
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(content))) > 0
}
