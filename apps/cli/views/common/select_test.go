// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common_test

import (
	"testing"

	"github.com/daytonaio/daytona/cli/views/common"
)

func TestSelectNonInteractive(t *testing.T) {
	forceNonInteractive(t)

	choice, err := common.Select("Choose an option", []common.SelectItem{
		{Title: "first", Desc: "first option"},
		{Title: "second", Desc: "second option"},
	})
	if choice != "" {
		t.Errorf("Select() in non-interactive mode returned %q, want empty string", choice)
	}
	assertUsageClierr(t, err, "re-run interactively or provide the value via flags")
}
