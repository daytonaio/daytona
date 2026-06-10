//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import "os"

// ownerGroup returns empty owner/group on Windows: file ownership lives in
// the security descriptor (GetNamedSecurityInfo), not in os.FileInfo, and the
// wire contract tolerates empty strings for these fields.
func ownerGroup(os.FileInfo) (owner, group string) {
	return "", ""
}
