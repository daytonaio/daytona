// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package creation_wizard

import (
	"log"

	"github.com/daytonaio/daytona/cli/cmd/views"

	"github.com/charmbracelet/huh"
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
