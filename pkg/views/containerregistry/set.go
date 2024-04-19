// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"errors"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

type RegistryView struct {
	Server   string
	Username string
	Password string
}

func RegistryCreationView(registryView *RegistryView, registries []serverapiclient.ContainerRegistry, editing bool) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Server URL").
				Value(&registryView.Server).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("server URL can not be blank")
					}
					return nil
				}),
			huh.NewInput().
				Title("Username").
				Value(&registryView.Username).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("username can not be blank")
					}
					return nil
				}),
			huh.NewInput().
				Title("Password").
				Password(true).
				Value(&registryView.Password).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("password can not be blank")
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
