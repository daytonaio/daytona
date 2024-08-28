// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"errors"
	"log"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	util "github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/huh"
)

type ProfileAddView struct {
	ProfileName string
	ApiUrl      string
	ApiKey      string
}

func ProfileCreationView(c *config.Config, profileAddView *ProfileAddView, editing bool) {
	nameInput := huh.NewInput().
		Title("Profile name").
		Validate(func(str string) error {
			if str == "" {
				return errors.New("profile name can not be blank")
			}
			result, err := util.GetValidatedName(str)
			if err != nil {
				return err
			}

			if !editing {
				for _, profile := range c.Profiles {
					if strings.EqualFold(profile.Name, str) {
						return errors.New("profile name already exists")
					}
				}
			}
            profileAddView.ProfileName = result
			return nil
		}).
		Value(&profileAddView.ProfileName)

	form := huh.NewForm(
		huh.NewGroup(
			nameInput,
			huh.NewInput().
				Title("Server API URL").
				Description("If you want to connect to a remote Daytona Server, start by running 'daytona api-key new' on the remote machine").
				Value(&profileAddView.ApiUrl).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("server API URL can not be blank")
					}

					_, err := url.ParseRequestURI(str)
					if err != nil {
						return errors.New("invalid url, must be of http/https format")
					}

					return nil
				}),
			huh.NewInput().
				Title("Server API Key").
				Password(true).
				Value(&profileAddView.ApiKey).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("server API Key can not be blank")
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
