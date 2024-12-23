// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/huh"
)

func RunnerRegistrationView(name *string, existingNames []string) error {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Runner Name").
				Value(name).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("name can not be blank")
					}
					for _, a := range existingNames {
						if a == str {
							return errors.New("name already in use")
						}
					}
					return nil
				}),
		),
	).WithHeight(5).WithTheme(views.GetCustomTheme()).Run()
}
