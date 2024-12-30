// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package purge

import (
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
)

func ConfirmPrompt(confirmCheck *bool) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you sure you want to clear all the data from Daytona?").
				Description("This action is irreversible.").
				Value(confirmCheck),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
