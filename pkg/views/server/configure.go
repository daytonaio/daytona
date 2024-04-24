// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
)

type ServerUpdateKeyView struct {
	GenerateNewKey   bool
	PathToPrivateKey string
}

func ConfigurationForm(config *serverapiclient.ServerConfig) *serverapiclient.ServerConfig {
	projectStartCommands := ""

	for _, command := range config.DefaultProjectPostStartCommands {
		projectStartCommands += strings.ReplaceAll(command, ",", "\\,") + ","
	}
	projectStartCommands = strings.TrimRight(projectStartCommands, ",")

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
				Description("Comma separated list of commands. To use ',' in commands, escape them like this '\\,'").
				Value(&projectStartCommands),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	startCommands := []string{}
	for _, command := range splitEscaped(projectStartCommands, ',') {
		startCommands = append(startCommands, strings.ReplaceAll(command, "\\,", ","))
	}

	config.DefaultProjectPostStartCommands = startCommands

	return config
}

func splitEscaped(s string, sep rune) []string {
	var result []string
	var builder strings.Builder
	escaping := false

	for _, c := range s {
		if c == '\\' && !escaping {
			escaping = true
			continue
		}

		if c == sep && !escaping {
			result = append(result, builder.String())
			builder.Reset()
			continue
		}

		builder.WriteRune(c)
		escaping = false
	}

	result = append(result, builder.String())
	return result
}
