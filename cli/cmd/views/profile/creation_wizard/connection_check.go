// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views_profile

import (
	"log"

	"github.com/daytonaio/daytona/cli/cmd/views"

	"github.com/charmbracelet/huh"
)

func IgnoreConnectionFailedCheck(ignoreCheck *bool, description string) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Server hasn't been set up. Add profile anyway?").
				Description(description).
				Value(ignoreCheck),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
