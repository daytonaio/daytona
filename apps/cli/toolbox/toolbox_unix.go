// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build unix

package toolbox

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// setupResizeHandler installs a SIGWINCH handler that forwards terminal resize
// events to the remote TTY session. Returns a cleanup function that stops
// signal handling.
func setupResizeHandler(ctx context.Context, proxyURL, sandboxId, sessionID string, c *Client, auth http.Header) func() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)

	go func() {
		for range sigChan {
			cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				c.resizePTYSession(ctx, proxyURL, sandboxId, sessionID, uint16(cols), uint16(rows), auth)
			}
		}
	}()

	return func() {
		signal.Stop(sigChan)
		close(sigChan)
	}
}
