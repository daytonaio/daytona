// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package purge

import (
	"fmt"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
)

func PurgeResourcesPrompt(continuePurge *bool, numOfTargets, numOfWorkspaces, numOfBuilds int) {
	titleMsg := fmt.Sprintf("Leftover resources found: [targets: %d, workspaces: %d, builds: %d]\nWould you like to continue with purge?", numOfTargets, numOfWorkspaces, numOfBuilds)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(titleMsg).
				Description("This action is irreversible.").
				Affirmative("Continue").
				Negative("Abort").
				Value(continuePurge),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
