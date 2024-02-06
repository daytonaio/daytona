// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"

	"github.com/charmbracelet/huh"
	log "github.com/sirupsen/logrus"
)

const workspaceKeyFileName = "workspace_key"
const defaultProjectBaseImage = "daytonaio/workspace-project:latest"

func Configure() error {
	defaultWorkspaceDir, err := getDefaultWorkspaceDir()
	if err != nil {
		return err
	}

	projectBaseImage := defaultProjectBaseImage
	defaultWorkspaceDirInput := defaultWorkspaceDir

	existingConfig, err := GetConfig()
	if err == nil && existingConfig != nil {
		projectBaseImage = existingConfig.ProjectBaseImage
		defaultWorkspaceDirInput = existingConfig.DefaultWorkspaceDir
	}

	if projectBaseImage == "" {
		projectBaseImage = defaultProjectBaseImage
	}

	if defaultWorkspaceDirInput == "" {
		defaultWorkspaceDirInput = defaultWorkspaceDir
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Base Image").
				Value(&projectBaseImage),
			huh.NewInput().
				Title("Default Workspace Directory").
				Value(&defaultWorkspaceDirInput).
				Validate(func(s string) error {
					_, err := os.Stat(s)
					if os.IsNotExist(err) {
						return os.MkdirAll(s, 0700)
					}

					return err
				}),
		),
	)

	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}

	c := Config{
		ProjectBaseImage:    projectBaseImage,
		DefaultWorkspaceDir: defaultWorkspaceDirInput,
	}

	return c.Save()
}
