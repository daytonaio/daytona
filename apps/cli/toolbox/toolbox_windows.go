// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build windows

package toolbox

import (
	"context"
	"net/http"
)

// setupResizeHandler is a no-op on Windows because SIGWINCH is not available.
// Terminal resize events are not forwarded on this platform.
func setupResizeHandler(_ context.Context, _, _, _ string, _ *Client, _ http.Header) func() {
	return func() {}
}
