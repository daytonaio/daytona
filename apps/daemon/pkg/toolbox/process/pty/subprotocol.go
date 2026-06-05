// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ptyEnvsSubprotocolPrefix carries PTY env vars as a WebSocket subprotocol token
// (not a header/query) so they reach the daemon across all SDK runtimes and stay
// out of request URLs.
const ptyEnvsSubprotocolPrefix = "X-Daytona-Pty-Envs~"

// extractPtyEnvsSubprotocol decodes env vars from the X-Daytona-Pty-Envs~ token
// (base64url JSON) in Sec-WebSocket-Protocol. Empty map if absent, error if malformed.
func extractPtyEnvsSubprotocol(header http.Header) (map[string]string, error) {
	envs := make(map[string]string)

	subprotocols := header.Get("Sec-WebSocket-Protocol")
	if subprotocols == "" {
		return envs, nil
	}

	for _, subprotocol := range strings.Split(subprotocols, ",") {
		subprotocol = strings.TrimSpace(subprotocol)
		if !strings.HasPrefix(subprotocol, ptyEnvsSubprotocolPrefix) {
			continue
		}

		encoded := strings.TrimPrefix(subprotocol, ptyEnvsSubprotocolPrefix)
		decoded, err := base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode PTY envs subprotocol: %w", err)
		}
		if err := json.Unmarshal(decoded, &envs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal PTY envs subprotocol: %w", err)
		}
		return envs, nil
	}

	return envs, nil
}
