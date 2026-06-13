// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import "fmt"

// BuildExtraHeaders constructs the extra header map for proxy target builders.
// X-Forwarded-Host is only set when the incoming request does not already carry
// one, preserving the upstream value per standard reverse-proxy convention.
func BuildExtraHeaders(apiKey, incomingXForwardedHost, requestHost string) map[string]string {
	extraHeaders := map[string]string{
		"X-Daytona-Authorization": fmt.Sprintf("Bearer %s", apiKey),
	}
	if incomingXForwardedHost == "" {
		extraHeaders["X-Forwarded-Host"] = requestHost
	}
	return extraHeaders
}
