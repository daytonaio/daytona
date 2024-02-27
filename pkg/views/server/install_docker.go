// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
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
