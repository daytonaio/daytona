// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views_profile

import (
	"log"

	"github.com/daytonaio/daytona/cli/cmd/views"

	"github.com/charmbracelet/huh"
)

func IgnoreConnectionFailedCheck(ignoreCheck *bool, description string) {
	errTheme := views.GetCustomTheme()
	errTheme.Focused.Title.Foreground(views.Red).Bold(true)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Connection to Daytona failed. Continue anyway?").
				Description(description).
				Value(ignoreCheck),
		),
	).WithTheme(errTheme)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
