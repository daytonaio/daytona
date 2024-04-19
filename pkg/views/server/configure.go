// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"log"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"
)

var CommandsInputHelp = "Comma separated list of commands. To use ',' in commands, escape them like this '\\,'"

type ServerUpdateKeyView struct {
	GenerateNewKey   bool
	PathToPrivateKey string
}

func ConfigurationForm(config *serverapiclient.ServerConfig) *serverapiclient.ServerConfig {
	projectStartCommands := util.GetJoinedCommands(config.DefaultProjectPostStartCommands)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Providers Directory").
				Value(config.ProvidersDir).
				Validate(func(s string) error {
					_, err := os.Stat(s)
					if os.IsNotExist(err) {
						return os.MkdirAll(s, 0700)
					}

					return err
				}),
			huh.NewInput().
				Title("Registry URL").
				Value(config.RegistryUrl),
			huh.NewInput().
				Title("Server Download URL").
				Value(config.ServerDownloadUrl),
			huh.NewInput().
				Title("Default Project Image").
				Value(config.DefaultProjectImage),
			huh.NewInput().
				Title("Default Project User").
				Value(config.DefaultProjectUser),
			huh.NewInput().
				Title("Default Project Post Start Commands").
				Description(CommandsInputHelp).
				Value(&projectStartCommands),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	config.DefaultProjectPostStartCommands = util.GetSplitCommands(projectStartCommands)

	return config
}
