// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/daytonaio/daytona/cli/internal/clierr"
)

// Confirm shows an interactive yes/no confirmation prompt and returns the
// user's choice. In non-interactive mode it fails instead of prompting.
func Confirm(question string) (bool, error) {
	if !internal.Interactive() {
		return false, clierr.New(clierr.CategoryUsage, "confirmation required in non-interactive mode").WithHint("pass --yes to proceed")
	}

	var confirmed bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(question).
				Value(&confirmed),
		).WithTheme(GetCustomTheme()),
	)

	if err := form.Run(); err != nil {
		return false, err
	}

	return confirmed, nil
}
