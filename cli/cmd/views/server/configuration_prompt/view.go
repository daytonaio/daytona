// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package configuration_prompt

import (
	"log"
	"os"

	"github.com/daytonaio/daytona/common/api_client"

	"github.com/charmbracelet/huh"
)

type ServerUpdateKeyView struct {
	GenerateNewKey   bool
	PathToPrivateKey string
}

func ConfigurationForm(config *api_client.ServerConfig) *api_client.ServerConfig {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Plugins Directory").
				Value(config.PluginsDir).
				Validate(func(s string) error {
					_, err := os.Stat(s)
					if os.IsNotExist(err) {
						return os.MkdirAll(s, 0700)
					}

					return err
				}),
			huh.NewInput().
				Title("Plugin Registry URL").
				Value(config.PluginRegistryUrl),
			huh.NewInput().
				Title("Server Download URL").
				Value(config.ServerDownloadUrl),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	return config
}
