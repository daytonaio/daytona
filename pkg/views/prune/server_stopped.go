// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prune

import (
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
)

func ServerStoppedPrompt(serverStoppedCheck *bool) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Please stop the Daytona Server before continuing").
				Description("Removing the config directory requires the Daytona Server to be stopped.").
				Affirmative("Continue").
				Negative("Abort").
				Value(serverStoppedCheck),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
