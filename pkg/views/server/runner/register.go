// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/huh"
)

func RunnerRegistrationView(alias *string, existingAliases []string) error {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Runner Alias").
				Value(alias).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("alias can not be blank")
					}
					for _, a := range existingAliases {
						if a == str {
							return errors.New("alias already in use")
						}
					}
					return nil
				}),
		),
	).WithHeight(5).WithTheme(views.GetCustomTheme()).Run()
}
