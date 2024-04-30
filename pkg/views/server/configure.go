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
	view_util "github.com/daytonaio/daytona/pkg/views/util"
)

var CommandsInputHelp = "Comma separated list of commands. To use ',' in commands, escape them like this '\\,'"

type ServerUpdateKeyView struct {
	GenerateNewKey   bool
	PathToPrivateKey string
}

func ConfigurationForm(config *serverapiclient.ServerConfig) *serverapiclient.ServerConfig {
	projectStartCommands := view_util.GetJoinedCommands(config.DefaultProjectPostStartCommands)
	apiPortView := strconv.Itoa(int(config.GetApiPort()))
	headscalePortView := strconv.Itoa(int(config.GetHeadscalePort()))
	frpsPortView := strconv.Itoa(int(config.Frps.GetPort()))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Providers Directory").
				Description("Directory will be created if it does not exist").
				Value(config.ProvidersDir).
				Validate(directoryValidator(config.ProvidersDir)),
			huh.NewInput().
				Title("Registry URL").
				Value(config.RegistryUrl),
			huh.NewInput().
				Title("Server Download URL").
				Value(config.ServerDownloadUrl),
		),
		huh.NewGroup(
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
		huh.NewGroup(
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
				Description("Directory will be created if it does not exist").
				Value(config.BinariesPath).
				Validate(directoryValidator(config.BinariesPath)),
			huh.NewInput().
				Title("Log File Path").
				Description("File will be created if it does not exist").
				Value(config.LogFilePath).
				Validate(func(s string) error {
					_, err := os.Stat(s)
					if os.IsNotExist(err) {
						_, err = os.Create(s)
					}

					return err
				}),
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

	config.DefaultProjectPostStartCommands = view_util.GetSplitCommands(projectStartCommands)

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

func directoryValidator(path *string) func(string) error {
	return func(string) error {
		_, err := os.Stat(*path)
		if os.IsNotExist(err) {
			return os.MkdirAll(*path, 0700)
		}

		return err
	}
}
