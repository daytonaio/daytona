// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/env/form"
)

func SetEnvVarsView(envVarsMap *map[string]string) error {
	return runSetEnvVarsForm(envVarsMap)
}

func runSetEnvVarsForm(envVarsMap *map[string]string) error {
	var key, value string

	m := form.NewFormModel(&key, &value)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		return err
	}

	if form.IsUserCancelled() {
		return fmt.Errorf("user cancelled")
	}

	(*envVarsMap)[key] = value

	var addAntoher bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add another environment variable?").
				Value(&addAntoher),
		),
	).WithTheme(views.GetCustomTheme()).WithHeight(6)

	err := form.Run()
	if err != nil {
		return err
	}

	if addAntoher {
		return runSetEnvVarsForm(envVarsMap)
	}

	return nil
}
