// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"net/http"
	"time"
)

// Defaults for ApplyServerTimeouts; overridable from main() before serving.
var (
	// Should sit above any upstream LB idle timeout so the LB closes first.
	ServerIdleTimeout = 70 * time.Second

	// Slowloris defense (gosec G112).
	ServerReadHeaderTimeout = 10 * time.Second
)

// ApplyServerTimeouts fills in zero-valued IdleTimeout and ReadHeaderTimeout
// on s with the package defaults.
func ApplyServerTimeouts(s *http.Server) {
	if s.IdleTimeout == 0 {
		s.IdleTimeout = ServerIdleTimeout
	}
	if s.ReadHeaderTimeout == 0 {
		s.ReadHeaderTimeout = ServerReadHeaderTimeout
	}
}
