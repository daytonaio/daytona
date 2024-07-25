// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

type ServerUpdateKeyView struct {
	GenerateNewKey   bool
	PathToPrivateKey string
}

func ConfigurationForm(config *apiclient.ServerConfig, containerRegistries []apiclient.ContainerRegistry) *apiclient.ServerConfig {
	apiPortView := strconv.Itoa(int(config.GetApiPort()))
	headscalePortView := strconv.Itoa(int(config.GetHeadscalePort()))
	frpsPortView := strconv.Itoa(int(config.Frps.GetPort()))
	localBuilderRegistryPort := strconv.Itoa(int(config.GetLocalBuilderRegistryPort()))

	builderContainerRegistryOptions := []huh.Option[string]{{
		Key:   "Local registry managed by Daytona",
		Value: "local",
	}}
	for _, cr := range containerRegistries {
		builderContainerRegistryOptions = append(builderContainerRegistryOptions, huh.Option[string]{Key: *cr.Server, Value: *cr.Server})
	}

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
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Builder Image").
				Description("Image dependencies: docker, @devcontainers/cli (node package)").
				Value(config.BuilderImage),
			huh.NewSelect[string]().
				Title("Builder Registry").
				Description("To add options, add a container registry with 'daytona cr set'").
				Options(
					builderContainerRegistryOptions...,
				).
				Value(config.BuilderRegistryServer),
			huh.NewInput().
				Title("Build Image Namespace").
				Description("Namespace to be used when tagging and pushing build images").
				Value(config.BuildImageNamespace),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Local Builder Registry Port").
				Value(&localBuilderRegistryPort).
				Validate(createPortValidator(config, &localBuilderRegistryPort, config.LocalBuilderRegistryPort)),
			huh.NewInput().
				Title("Local Builder Registry Image").
				Value(config.RegistryImage),
		).WithHideFunc(func() bool {
			return config.BuilderRegistryServer == nil || *config.BuilderRegistryServer != "local"
		}),
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
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func createPortValidator(config *apiclient.ServerConfig, portView *string, port *int32) func(string) error {
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
