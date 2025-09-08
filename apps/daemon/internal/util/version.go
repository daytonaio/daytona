// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	semver "github.com/Masterminds/semver/v3"
)

// ExtractSdkVersionFromHeader extracts the SDK version from the headers.
// If the X-Daytona-SDK-Version header is not present, it looks through
// the Sec-WebSocket-Protocol header looking for the version protocol formatted like
// X-Daytona-SDK-Version/<version>.
// If no version is found, it returns an empty string.
func ExtractSdkVersionFromHeader(header http.Header) string {
	if v := header.Get("X-Daytona-SDK-Version"); v != "" {
		return v
	}

	// no explicit header; look through Sec-WebSocket-Protocol entries
	protocols := header.Get("Sec-WebSocket-Protocol")
	if protocols == "" {
		return ""
	}
	const prefix = "X-Daytona-SDK-Version~"
	// split comma-separated protocols
	for _, protocol := range strings.Split(protocols, ",") {
		protocol = strings.TrimSpace(protocol)
		if strings.HasPrefix(protocol, prefix) {
			// found version protocol; split off the version
			parts := strings.SplitN(protocol, "~", 2)
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}

	return ""
}

// CompareVersions compares two versions and returns:
// 1 if v1 is greater than v2
// -1 if v1 is less than v2
// 0 if they are equal
//
// It considers pre-releases to be invalid if the ranges does not include one.
// If you want to have it include pre-releases a simple solution is to include -0 in your range.
func CompareVersions(v1 string, v2 string) (*int, error) {
	semverV1, err := semver.NewVersion(normalizeSemver(v1))
	if err != nil {
		return nil, fmt.Errorf("failed to parse semver v1: %s, normalized: %s, error: %w", v1, normalizeSemver(v1), err)
	}
	semverV2, err := semver.NewVersion(normalizeSemver(v2))
	if err != nil {
		return nil, fmt.Errorf("failed to parse semver v2: %s, normalized: %s, error: %w", v2, normalizeSemver(v2), err)
	}

	comparison := semverV1.Compare(semverV2)
	return &comparison, nil
}

func normalizeSemver(input string) string {
	// If it's already in the form X.Y.Z-suffix, return as-is.
	reAlreadyDashed := regexp.MustCompile(`^\d+\.\d+\.\d+-\S+$`)
	if reAlreadyDashed.MatchString(input) {
		return input
	}

	// If there's a non-digit suffix immediately after X.Y.Z, dash it.
	reNeedsDash := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(\D.+)$`)
	if reNeedsDash.MatchString(input) {
		return reNeedsDash.ReplaceAllString(input, `$1.$2.$3-$4`)
	}

	// Otherwise (pure X.Y.Z or something else), leave unchanged.
	return input
}
