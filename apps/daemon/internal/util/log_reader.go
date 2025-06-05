// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"bufio"
	"context"
	"io"
	"time"
)

func ReadLog(ctx context.Context, logReader io.Reader, follow bool, c chan []byte, errChan chan error) {
	reader := bufio.NewReader(logReader)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			bytes := make([]byte, 1024)
			_, err := reader.Read(bytes)
			if err != nil {
				if err != io.EOF {
					errChan <- err
					return
				} else if !follow {
					errChan <- io.EOF
					return
				}
				// Sleep for a short time to avoid busy-waiting
				time.Sleep(20 * time.Millisecond)
				continue
			}
			c <- bytes
		}
	}
}
