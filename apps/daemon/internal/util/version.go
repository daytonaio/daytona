// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"fmt"
	"regexp"

	semver "github.com/Masterminds/semver/v3"
)

func CompareVersions(v1 string, v2 string) (int, error) {
	semverV1, err := semver.NewVersion(normalizeSemver(v1))
	if err != nil {
		return 0, fmt.Errorf("failed to parse semver v1: %s, %s, error: %w", v1, normalizeSemver(v1), err)
	}
	semverV2, err := semver.NewVersion(normalizeSemver(v2))
	if err != nil {
		return 0, fmt.Errorf("failed to parse semver v2: %s, %s, error: %w", v2, normalizeSemver(v2), err)
	}

	return semverV1.Compare(semverV2), nil
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
