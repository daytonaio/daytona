// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package configuration_prompt

import (
	"log"
	"os"

	"github.com/daytonaio/daytona/common/grpc/proto/types"

	"github.com/charmbracelet/huh"
)

type ServerUpdateKeyView struct {
	GenerateNewKey   bool
	PathToPrivateKey string
}

func ConfigurationForm(config *types.ServerConfig) *types.ServerConfig {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Base Image").
				Value(&config.ProjectBaseImage),
			huh.NewInput().
				Title("Default Workspace Directory").
				Value(&config.DefaultWorkspaceDir).
				Validate(func(s string) error {
					_, err := os.Stat(s)
					if os.IsNotExist(err) {
						return os.MkdirAll(s, 0700)
					}

					return err
				}),
			huh.NewInput().
				Title("Plugins Directory").
				Value(&config.PluginsDir).
				Validate(func(s string) error {
					_, err := os.Stat(s)
					if os.IsNotExist(err) {
						return os.MkdirAll(s, 0700)
					}

					return err
				}),
			huh.NewInput().
				Title("Plugin Registry URL").
				Value(&config.PluginRegistryUrl),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	return config
}
