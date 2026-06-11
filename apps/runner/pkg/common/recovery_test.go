// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"testing"

	"github.com/daytonaio/runner/pkg/models"
)

// TestDeduceRecoveryType_FdExhaustionNotActionable guards the invariant that
// file-descriptor exhaustion signatures are surfacing-only (degraded.go) and
// can never reach RecoverSandbox/RecoverFromStorageLimit through the
// recoverable-error classification.
func TestDeduceRecoveryType_FdExhaustionNotActionable(t *testing.T) {
	signatures := []string{
		"EMFILE: too many open files",
		"accept4: too many open files",
		"fork/exec /bin/sh: too many open files",
		"too many open files in system",
		"error while loading shared libraries: libc.so.6: cannot open shared object file: Error 24",
	}

	for _, msg := range signatures {
		if got := DeduceRecoveryType(msg); got != models.UnknownRecoveryType {
			t.Errorf("DeduceRecoveryType(%q) = %v, want UnknownRecoveryType", msg, got)
		}
		if IsRecoverable(msg) {
			t.Errorf("IsRecoverable(%q) = true, want false", msg)
		}
	}
}

// TestDeduceRecoveryType_StorageExpansion is a regression guard that the
// actionable storage-expansion path stays intact.
func TestDeduceRecoveryType_StorageExpansion(t *testing.T) {
	for _, msg := range []string{
		"no space left on device",
		"disk quota exceeded",
	} {
		if got := DeduceRecoveryType(msg); got != models.RecoveryTypeStorageExpansion {
			t.Errorf("DeduceRecoveryType(%q) = %v, want RecoveryTypeStorageExpansion", msg, got)
		}
		if !IsRecoverable(msg) {
			t.Errorf("IsRecoverable(%q) = false, want true", msg)
		}
	}
}
