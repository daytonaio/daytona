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
	apiPortView := strconv.Itoa(int(config.GetApiPort()))
	headscalePortView := strconv.Itoa(int(config.GetHeadscalePort()))
	frpsPortView := strconv.Itoa(int(config.Frps.GetPort()))

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
				Value(&apiPortView).
				Validate(createPortValidator(config, &apiPortView, config.ApiPort)),
			huh.NewInput().
				Title("Headscale Port").
				Value(&headscalePortView).
				Validate(createPortValidator(config, &headscalePortView, config.HeadscalePort)),
			huh.NewInput().
				Title("Binaries Path").
				Value(config.BinariesPath),
			huh.NewInput().
				Title("Targets File Path").
				Value(config.TargetsFilePath),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Frps Domain").
				Value(config.Frps.Domain),
			huh.NewInput().
				Title("Frps Port").
				Value(&frpsPortView).
				Validate(createPortValidator(config, &frpsPortView, config.Frps.Port)),
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

func createPortValidator(config *serverapiclient.ServerConfig, portView *string, port *int32) func(string) error {
	return func(string) error {
		validatePort, err := strconv.Atoi(*portView)
		if err != nil {
			return errors.New("failed to parse port")
		}
		if validatePort < 0 || validatePort > 65535 {
			return errors.New("port out of range")
		}
		*port = int32(validatePort)

		if *config.ApiPort == *config.HeadscalePort {
			return errors.New("port conflict")
		}

		return nil
	}
}
