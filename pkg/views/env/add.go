// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
)

func AddEnvVarsView(envVarsMap *map[string]string) error {
	var addAntoher bool

	var key string
	var value string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Key").
				Value(&key),
			huh.NewInput().
				Title("Value").
				Value(&value),
		),
	).WithTheme(views.GetCustomTheme()).WithHeight(12)

	err := form.Run()
	if err != nil {
		return err
	}

	(*envVarsMap)[key] = value

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add another environment variable?").
				Value(&addAntoher),
		),
	).WithTheme(views.GetCustomTheme()).WithHeight(12)

	err = form.Run()
	if err != nil {
		return err
	}

	if addAntoher {
		return AddEnvVarsView(envVarsMap)
	}

	return nil
}
