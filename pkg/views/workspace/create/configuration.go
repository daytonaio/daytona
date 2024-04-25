// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	configure "github.com/daytonaio/daytona/pkg/views/server"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

var configurationHelpLine = lipgloss.NewStyle().Foreground(views.Gray).Render("enter: next  f10: advanced configuration")

type ProjectConfigurationData struct {
	BuilderChoice        string
	DevcontainerFilePath string
	Image                string
	User                 string
	PostStartCommands    []string
	EnvVars              map[string]string
}

func ConfigureProjects(projectList []serverapiclient.CreateWorkspaceRequestProject, apiServerConfig serverapiclient.ServerConfig) ([]serverapiclient.CreateWorkspaceRequestProject, error) {
	var currentProject *serverapiclient.CreateWorkspaceRequestProject

	if len(projectList) > 1 {
		currentProject = selection.GetProjectRequestFromPrompt(projectList)
		if currentProject == nil {
			return projectList, errors.New("project is required")
		}
	} else {
		currentProject = &projectList[0]
	}

	if currentProject.Name == selection.DoneConfiguring.Name {
		return projectList, nil
	}

	devContainerFilePath := ".devcontainer/devcontainer.json"
	builderChoice := "auto"

	if currentProject.Build != nil {
		if currentProject.Build.Devcontainer != nil {
			builderChoice = "devcontainer"
			devContainerFilePath = *currentProject.Build.Devcontainer.DevContainerFilePath
		}
	} else {
		builderChoice = "none"
	}

	projectConfigurationData := &ProjectConfigurationData{
		BuilderChoice:        builderChoice,
		DevcontainerFilePath: devContainerFilePath,
		Image:                *currentProject.Image,
		User:                 *currentProject.User,
		PostStartCommands:    currentProject.PostStartCommands,
		EnvVars:              *currentProject.EnvVars,
	}

	form := GetProjectConfigurationForm(projectConfigurationData)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	for i := range projectList {
		if projectList[i].Name == currentProject.Name {
			if projectConfigurationData.BuilderChoice == "none" {
				projectList[i].Image = apiServerConfig.DefaultProjectImage
				projectList[i].User = apiServerConfig.DefaultProjectUser
				continue
			}

			projectList[i].Image = &projectConfigurationData.Image
			projectList[i].User = &projectConfigurationData.User
			projectList[i].PostStartCommands = projectConfigurationData.PostStartCommands
			projectList[i].EnvVars = &projectConfigurationData.EnvVars

			if projectConfigurationData.BuilderChoice == "auto" {
				projectList[i].Build = &serverapiclient.ProjectBuild{}
				continue
			}

			if projectConfigurationData.BuilderChoice == "devcontainer" {
				projectList[i].Build = &serverapiclient.ProjectBuild{
					Devcontainer: &serverapiclient.ProjectBuildDevcontainer{
						DevContainerFilePath: &projectConfigurationData.DevcontainerFilePath,
					},
				}
			}
		}
	}

	if len(projectList) == 1 {
		return projectList, nil
	}

	return ConfigureProjects(projectList, apiServerConfig)
}

func GetProjectConfigurationForm(projectConfiguration *ProjectConfigurationData) *huh.Form {
	var buildOptions []huh.Option[string]
	// TODO: move to variable
	buildOptions = append(buildOptions, huh.Option[string]{Key: "Automatic", Value: "auto"})
	buildOptions = append(buildOptions, huh.Option[string]{Key: "Devcontainer", Value: "devcontainer"})
	buildOptions = append(buildOptions, huh.Option[string]{Key: "Custom image", Value: "custom-image"})
	buildOptions = append(buildOptions, huh.Option[string]{Key: "None", Value: "none"})

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a build configuration").
				Options(
					buildOptions...,
				).
				Value(&projectConfiguration.BuilderChoice),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Custom container image").
				Value(&projectConfiguration.Image),
			huh.NewInput().
				Title("Container user").
				Value(&projectConfiguration.User),
		).WithHideFunc(func() bool {
			return projectConfiguration.BuilderChoice != "custom-image"
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Devcontainer file path").
				Value(&projectConfiguration.DevcontainerFilePath).Validate(func(s string) error {
				if s == "" {
					return errors.New("devcontainer file path is required")
				}
				return nil
			}),
		).WithHideFunc(func() bool {
			return projectConfiguration.BuilderChoice != "devcontainer"
		}),
		huh.NewGroup(
			configure.GetPostStartCommandsInput(&projectConfiguration.PostStartCommands, "Post start commands"),
		).WithHideFunc(func() bool {
			return projectConfiguration.BuilderChoice == "devcontainer"
		}),
		huh.NewGroup(
			views.GetEnvVarsInput(&projectConfiguration.EnvVars),
		),
	).WithTheme(views.GetCustomTheme())

	keyMap := huh.NewDefaultKeyMap()
	keyMap.Text = huh.TextKeyMap{
		NewLine: key.NewBinding(key.WithKeys("alt+enter"), key.WithHelp("alt+enter", "new line")),
		Next:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next")),
		Prev:    key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev")),
	}

	form = form.WithKeyMap(keyMap)

	return form
}
