// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile_create_wizard

import (
	"log"

	"github.com/daytonaio/daytona/cmd/views"

	"github.com/charmbracelet/huh"
)

func DockerPrompt(dockerCheck *bool) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Install Docker?").
				Description("Docker was not found on the remote machine.").
				Value(dockerCheck),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
