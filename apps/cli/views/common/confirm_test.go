// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common_test

import (
	"errors"
	"testing"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	"github.com/daytonaio/daytona/cli/views/common"
)

// forceNonInteractive makes internal.Interactive() return false for the
// duration of the test, regardless of the test process TTY.
func forceNonInteractive(t *testing.T) {
	t.Helper()
	prev := internal.NoInput
	internal.NoInput = true
	t.Cleanup(func() { internal.NoInput = prev })
}

// assertUsageClierr fails the test unless err is a *clierr.Error with the
// usage category and the expected hint.
func assertUsageClierr(t *testing.T, err error, wantHint string) {
	t.Helper()
	var cliErr *clierr.Error
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *clierr.Error, got %T: %v", err, err)
	}
	if cliErr.Category != clierr.CategoryUsage {
		t.Errorf("category = %q, want %q", cliErr.Category, clierr.CategoryUsage)
	}
	if cliErr.Hint != wantHint {
		t.Errorf("hint = %q, want %q", cliErr.Hint, wantHint)
	}
}

func TestConfirmNonInteractive(t *testing.T) {
	forceNonInteractive(t)

	confirmed, err := common.Confirm("Delete 2 sandboxes?")
	if confirmed {
		t.Error("Confirm() in non-interactive mode returned true, want false")
	}
	assertUsageClierr(t, err, "pass --yes to proceed")
}
