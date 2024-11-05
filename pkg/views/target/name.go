// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"errors"
	"fmt"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
)

func SetTargetNameView(name *string, existingNames []string) {
	fmt.Println(existingNames)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Value(name).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("name can not be blank")
					}
					for _, existingName := range existingNames {
						if existingName == str {
							return errors.New("name already in use")
						}
					}
					return nil
				}),
		),
	).WithHeight(5).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
