// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"net/url"
	"strings"
)

// ExtractHost extracts the host from a full URL
func ExtractHost(fullURL string) string {
	u, err := url.Parse(fullURL)
	if err != nil {
		return ""
	}
	return u.Host
}

// ExtractScheme extracts the scheme (http/https) from a full URL
func ExtractScheme(fullURL string) string {
	u, err := url.Parse(fullURL)
	if err != nil {
		// Default to https if parse fails
		if strings.HasPrefix(fullURL, "http://") {
			return "http"
		}
		return "https"
	}
	return u.Scheme
}

// ExtractPath extracts the path from a full URL
func ExtractPath(fullURL string) string {
	u, err := url.Parse(fullURL)
	if err != nil {
		return ""
	}
	return u.Path
}

// ConvertToWebSocketURL converts an HTTP(S) URL to WS(S)
func ConvertToWebSocketURL(httpURL string) string {
	wsURL := strings.Replace(httpURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	return wsURL
}
