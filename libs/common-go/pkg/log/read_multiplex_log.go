// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"bufio"
	"bytes"
	"context"
	"io"
)

// Stream prefixes for multiplexing stdout/stderr in logs
var (
	STDOUT_PREFIX = []byte{0x01, 0x01, 0x01}
	STDERR_PREFIX = []byte{0x02, 0x02, 0x02}
)

func ReadMultiplexedLog(ctx context.Context, logReader io.Reader, follow bool, stdoutChan chan []byte, stderrChan chan []byte, errChan chan error) {
	reader := bufio.NewReader(logReader)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			b := make([]byte, 1024)
			_, err := reader.Read(b)
			if err != nil {
				if err != io.EOF {
					errChan <- err
				} else if !follow {
					errChan <- io.EOF
					return
				}
				continue
			}
			if bytes.HasPrefix(b, STDOUT_PREFIX) {
				stdoutChan <- b[len(STDOUT_PREFIX):]
			} else if bytes.HasPrefix(b, STDERR_PREFIX) {
				stderrChan <- b[len(STDERR_PREFIX):]
			}
		}
	}
}
