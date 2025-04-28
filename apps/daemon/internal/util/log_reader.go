// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bufio"
	"context"
	"io"
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
				} else if !follow {
					errChan <- io.EOF
					return
				}
				continue
			}
			c <- bytes
		}
	}
}
