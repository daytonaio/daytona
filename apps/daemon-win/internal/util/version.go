// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// ExtractSdkVersionFromHeader extracts the SDK version from HTTP headers
func ExtractSdkVersionFromHeader(header http.Header) string {
	userAgent := header.Get("User-Agent")
	if userAgent == "" {
		return ""
	}

	// Look for version pattern in User-Agent
	parts := strings.Split(userAgent, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}

// CompareVersions compares two semantic versions
// Returns:
//   - negative if v1 < v2
//   - zero if v1 == v2
//   - positive if v1 > v2
//   - nil if either version is invalid
func CompareVersions(v1, v2 string) (*int, error) {
	if v1 == "" || v2 == "" {
		return nil, nil
	}

	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	semVer1, err := semver.NewVersion(v1)
	if err != nil {
		return nil, err
	}

	semVer2, err := semver.NewVersion(v2)
	if err != nil {
		return nil, err
	}

	result := semVer1.Compare(semVer2)
	return &result, nil
}
