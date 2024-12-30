// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
)

func ConfirmPrompt(confirmCheck *bool) {
	views.RenderInfoMessageBold("This commands registers and starts the Daytona Server daemon.\nFor running the Server in the current terminal session use 'daytona serve'.")

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(" Do you want to continue?").
				Value(confirmCheck),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
