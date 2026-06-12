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
// names and unsafe to embed in paths and logs. The allowlist also excludes
// path separators (/ and \), so no separate separator check is needed.
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

	// Reject ".." anywhere, not just a leading dot: defense-in-depth against
	// traversal, and the recording dashboard's serveVideo 403s any path
	// containing "..", so a recording named after such a label would be
	// accepted here but permanently unplayable there.
	if strings.Contains(label, "..") {
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
