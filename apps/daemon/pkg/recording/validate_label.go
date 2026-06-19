// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"regexp"
	"strings"
)

// labelPattern restricts labels to characters that are safe to embed in a
// file name on every supported OS. Only plain spaces are allowed — not the
// full \s class (tab/newline/CR/FF/VT), which is invalid in Windows file
// names and unsafe to embed in paths and logs.
var labelPattern = regexp.MustCompile(`^[A-Za-z0-9. _-]+$`)

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

	if !labelPattern.MatchString(label) {
		return ErrInvalidLabel
	}

	return nil
}
