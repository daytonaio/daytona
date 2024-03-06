// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"errors"
	"log"
	"regexp"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/huh"
)

type ProfileAddView struct {
	ProfileName string
	ApiUrl      string
}

func ProfileCreationView(c *config.Config, profileAddView *ProfileAddView, editing bool) {
	nameInput := huh.NewInput().
		Title("Profile name").
		Validate(func(str string) error {
			if str == "" {
				return errors.New("profile name can not be blank")
			}
			if match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", str); !match {
				return errors.New("profile name must be alphanumeric only")
			}

			if !editing {
				for _, profile := range c.Profiles {
					if strings.EqualFold(profile.Name, str) {
						return errors.New("profile name already exists")
					}
				}
			}

			return nil
		}).
		Value(&profileAddView.ProfileName)

	form := huh.NewForm(
		huh.NewGroup(
			nameInput,
			huh.NewInput().
				Title("Server API URL").
				Value(&profileAddView.ApiUrl).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("server API URL can not be blank")
					}
					return nil
				}),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
}
