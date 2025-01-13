// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package purge

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
)

func PurgeResourcesPrompt(continuePurge *bool, numOfTargets, numOfWorkspaces, numOfBuilds int) []string {
	commands := []string{}
	resources := []string{}

	if numOfBuilds > 0 {
		resources = append(resources, fmt.Sprintf("builds: %d", numOfBuilds))
		commands = append(commands, lipgloss.NewStyle().Foreground(views.DimmedGreen).Render("daytona build delete -af"))
	}

	if numOfWorkspaces > 0 {
		resources = append(resources, fmt.Sprintf("workspaces: %d", numOfWorkspaces))
		commands = append(commands, lipgloss.NewStyle().Foreground(views.DimmedGreen).Render("daytona delete -afy"))
	}

	if numOfTargets > 0 {
		resources = append(resources, fmt.Sprintf("targets: %d", numOfTargets))
		commands = append(commands, lipgloss.NewStyle().Foreground(views.DimmedGreen).Render("daytona target delete -afy"))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Leftover resources found: [%s]\nWould you like to continue with purge?", strings.Join(resources, ", "))).
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

	return commands
}
