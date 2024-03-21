// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
)

type ServerUpdateKeyView struct {
	GenerateNewKey   bool
	PathToPrivateKey string
}

func ConfigurationForm(config *serverapiclient.ServerConfig) *serverapiclient.ServerConfig {
	apiPort := strconv.Itoa(int(config.GetApiPort()))
	headscalePort := strconv.Itoa(int(config.GetHeadscalePort()))
	frpsPort := strconv.Itoa(int(config.Frps.GetPort()))

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
				Title("API Port").
				Value(&apiPort).
				Validate(createPortValidator(config, &apiPort, "api")),
			huh.NewInput().
				Title("Headscale Port").
				Value(&headscalePort).
				Validate(createPortValidator(config, &headscalePort, "headscale")),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Frps Domain").
				Value(config.Frps.Domain),
			huh.NewInput().
				Title("Frps Port").
				Value(&frpsPort).
				Validate(createPortValidator(config, &frpsPort, "frps")),
			huh.NewInput().
				Title("Frps Protocol").
				Value(config.Frps.Protocol),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func createPortValidator(config *serverapiclient.ServerConfig, s *string, field string) func(string) error {
	return func(string) error {
		port, err := strconv.Atoi(*s)
		if err != nil {
			return errors.New("failed to parse port")
		}
		if port < 0 || port > 65535 {
			return errors.New("port out of range")
		}
		validatedPort := int32(port)

		switch field {
		case "api":
			config.ApiPort = &validatedPort
		case "headscale":
			config.HeadscalePort = &validatedPort
		case "frps":
			config.Frps.Port = &validatedPort
		}

		if *config.ApiPort == *config.HeadscalePort {
			return errors.New("port conflict")
		}

		return nil
	}
}
