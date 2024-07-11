// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

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

	if config.IpWithProtocol == nil {
		config.IpWithProtocol = new(string)
	}

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
			GetPostStartCommandsInput(&config.DefaultProjectPostStartCommands, "Default Project Post Start Commands"),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Server IP with Protocol").
				// TODO: Add docs entry link
				Description("If the server has a static IP address, you can set it here to avoid using our reverse proxy\nNOTE: Due to current limitations of our Headscale server, the IP must be accessible over https").
				Value(config.IpWithProtocol).
				Validate(func(s string) error {
					if s == "" {
						return nil
					}
					if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
						return errors.New("invalid protocol")
					}
					return nil
				}),
			huh.NewInput().
				Title("API Port").
				Value(&apiPortView).
				Validate(createPortValidator(config, &apiPortView, config.ApiPort)),
			huh.NewInput().
				Title("Headscale Port").
				Value(&headscalePortView).
				Validate(createPortValidator(config, &headscalePortView, config.HeadscalePort)),
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
		).WithHideFunc(func() bool {
			return config.BuilderRegistryServer == nil || *config.BuilderRegistryServer != "local"
		}),
		huh.NewGroup(
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
		).Description("Frps is a reverse proxy server that is used for connecting workspaces to the server and public port forwarding"),
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

func GetPostStartCommandsInput(postStartCommands *[]string, title string) *huh.Text {
	postStartCommandsString := ""
	for _, command := range *postStartCommands {
		postStartCommandsString += command + "\n"
	}
	postStartCommandsString = strings.TrimSuffix(postStartCommandsString, "\n")

	return huh.NewText().
		Title(title).
		Description("Enter one command per line.").
		Value(&postStartCommandsString).
		Validate(func(s string) error {
			*postStartCommands = []string{}
			for _, line := range strings.Split(s, "\n") {
				if line == "" {
					continue
				}
				*postStartCommands = append(*postStartCommands, line)
			}

			return nil
		})
}
