// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
)

func RestartPrompt(restartCheck *bool) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Restart the server?").
				Description("The server needs to be restarted to complete the installation.").
				Value(restartCheck),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
