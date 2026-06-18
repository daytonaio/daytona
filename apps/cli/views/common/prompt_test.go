// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common_test

import (
	"testing"

	"github.com/daytonaio/daytona/cli/views/common"
)

func TestPromptForInputNonInteractive(t *testing.T) {
	forceNonInteractive(t)

	value, err := common.PromptForInput("prompt", "Title", "Description")
	if value != "" {
		t.Errorf("PromptForInput() in non-interactive mode returned %q, want empty string", value)
	}
	assertUsageClierr(t, err, "re-run interactively or provide the value via flags")
}
