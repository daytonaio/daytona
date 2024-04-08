// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"time"
)

func ReadLog(ctx context.Context, logReader *io.Reader, follow bool, c chan []byte, errChan chan error) {
	reader := bufio.NewReader(*logReader)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					errChan <- err
				} else if !follow {
					errChan <- io.EOF
					return
				}
				time.Sleep(500 * time.Millisecond) // Sleep to avoid busy loop
				continue
			}
			// Trim the newline character
			line = bytes.TrimRight(line, "\n")
			c <- line
		}
	}
}
