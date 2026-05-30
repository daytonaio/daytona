// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"regexp"
	"strings"
)

// validateLabel validates a user-provided label to prevent path injection
// and ensure it's safe for use in a filename. Returns error if invalid.
func validateLabel(label string) error {
	const maxLabelLength = 100

	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		return ErrInvalidLabel
	}

	if len(label) > maxLabelLength {
		return ErrInvalidLabel
	}

	if strings.Contains(label, "/") || strings.Contains(label, "\\") {
		return ErrInvalidLabel
	}

	if strings.HasPrefix(trimmed, ".") {
		return ErrInvalidLabel
	}

	safePattern := regexp.MustCompile(`^[A-Za-z0-9.\s_-]+$`)
	if !safePattern.MatchString(label) {
		return ErrInvalidLabel
	}

	return nil
}
