// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"net/http"
	"time"
)

const (
	// Must sit above any upstream LB/client idle timeout so the peer closes
	// idle connections first and never reuses one we already closed.
	serverIdleTimeout = 70 * time.Second

	// Slowloris defense (gosec G112).
	serverReadHeaderTimeout = 10 * time.Second
)

// ApplyServerTimeouts fills in zero-valued IdleTimeout and ReadHeaderTimeout
// on s with the package defaults. Callers that need different values can set
// the fields on s before calling.
func ApplyServerTimeouts(s *http.Server) {
	if s.IdleTimeout == 0 {
		s.IdleTimeout = serverIdleTimeout
	}
	if s.ReadHeaderTimeout == 0 {
		s.ReadHeaderTimeout = serverReadHeaderTimeout
	}
}
