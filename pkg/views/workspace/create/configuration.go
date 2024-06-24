// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	configure "github.com/daytonaio/daytona/pkg/views/server"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

const (
	DEVCONTAINER_FILEPATH = ".devcontainer/devcontainer.json"
)

var configurationHelpLine = lipgloss.NewStyle().Foreground(views.Gray).Render("enter: next  f10: advanced configuration")

type ProjectConfigurationData struct {
	BuildChoice          string
	DevcontainerFilePath string
	Image                string
	User                 string
	PostStartCommands    []string
	EnvVars              map[string]string
}

func NewProjectConfigurationData(buildChoice BuildChoice, devContainerFilePath string, currentProject *apiclient.CreateWorkspaceRequestProject, defaults *ProjectDefaults) *ProjectConfigurationData {
	image := *defaults.Image
	user := *defaults.ImageUser
	commands := []string{}
	envVars := map[string]string{}

	if currentProject.Image != nil {
		image = *currentProject.Image
	}

	if currentProject.User != nil {
		user = *currentProject.User
	}

	if currentProject.PostStartCommands != nil {
		commands = currentProject.PostStartCommands
	}

	if currentProject.EnvVars != nil {
		envVars = *currentProject.EnvVars
	}

	return &ProjectConfigurationData{
		BuildChoice:          string(buildChoice),
		DevcontainerFilePath: devContainerFilePath,
		Image:                image,
		User:                 user,
		PostStartCommands:    commands,
		EnvVars:              envVars,
	}
}

func ConfigureProjects(projectList *[]apiclient.CreateWorkspaceRequestProject, defaults ProjectDefaults) (bool, error) {
	var currentProject *apiclient.CreateWorkspaceRequestProject

	if len(*projectList) > 1 {
		currentProject = selection.GetProjectRequestFromPrompt(projectList)
		if currentProject == nil {
			return false, errors.New("project is required")
		}
	} else {
		currentProject = &((*projectList)[0])
	}

	if currentProject.Name == selection.DoneConfiguring.Name {
		return false, nil
	}

	devContainerFilePath := DEVCONTAINER_FILEPATH
	builderChoice := AUTOMATIC

	if currentProject.Build != nil {
		if currentProject.Build.Devcontainer != nil {
			builderChoice = DEVCONTAINER
			devContainerFilePath = *currentProject.Build.Devcontainer.DevContainerFilePath
		}
	} else {
		if *currentProject.Image == *defaults.Image && *currentProject.User == *defaults.ImageUser {
			builderChoice = NONE
		} else {
			builderChoice = CUSTOMIMAGE
		}
	}

	projectConfigurationData := NewProjectConfigurationData(builderChoice, devContainerFilePath, currentProject, &defaults)

	form := GetProjectConfigurationForm(projectConfigurationData)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	for i := range *projectList {
		if (*projectList)[i].Name == currentProject.Name {
			if projectConfigurationData.BuildChoice == string(NONE) {
				(*projectList)[i].Build = nil
				(*projectList)[i].Image = defaults.Image
				(*projectList)[i].User = defaults.ImageUser
				(*projectList)[i].PostStartCommands = projectConfigurationData.PostStartCommands
			}

			if projectConfigurationData.BuildChoice == string(CUSTOMIMAGE) {
				(*projectList)[i].Build = nil
				(*projectList)[i].Image = &projectConfigurationData.Image
				(*projectList)[i].User = &projectConfigurationData.User
				(*projectList)[i].PostStartCommands = projectConfigurationData.PostStartCommands
			}

			if projectConfigurationData.BuildChoice == string(AUTOMATIC) {
				(*projectList)[i].Build = &apiclient.ProjectBuild{}
				(*projectList)[i].Image = defaults.Image
				(*projectList)[i].User = defaults.ImageUser
				(*projectList)[i].PostStartCommands = projectConfigurationData.PostStartCommands
			}

			if projectConfigurationData.BuildChoice == string(DEVCONTAINER) {
				(*projectList)[i].Build = &apiclient.ProjectBuild{
					Devcontainer: &apiclient.ProjectBuildDevcontainer{
						DevContainerFilePath: &projectConfigurationData.DevcontainerFilePath,
					},
				}
				(*projectList)[i].Image = nil
				(*projectList)[i].User = nil
				(*projectList)[i].PostStartCommands = nil
			}

			(*projectList)[i].EnvVars = &projectConfigurationData.EnvVars
		}
	}

	if len(*projectList) == 1 {
		return true, nil
	}

	return ConfigureProjects(projectList, defaults)
}

func GetProjectConfigurationForm(projectConfiguration *ProjectConfigurationData) *huh.Form {
	buildOptions := []huh.Option[string]{
		{Key: "Automatic", Value: string(AUTOMATIC)},
		{Key: "Devcontainer", Value: string(DEVCONTAINER)},
		{Key: "Custom image", Value: string(CUSTOMIMAGE)},
		{Key: "None", Value: string(NONE)},
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a build configuration").
				Options(
					buildOptions...,
				).
				Value(&projectConfiguration.BuildChoice),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Custom container image").
				Value(&projectConfiguration.Image),
			huh.NewInput().
				Title("Container user").
				Value(&projectConfiguration.User),
		).WithHideFunc(func() bool {
			return projectConfiguration.BuildChoice != string(CUSTOMIMAGE)
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
			return projectConfiguration.BuildChoice != string(DEVCONTAINER)
		}),
		huh.NewGroup(
			configure.GetPostStartCommandsInput(&projectConfiguration.PostStartCommands, "Post start commands"),
		).WithHideFunc(func() bool {
			return projectConfiguration.BuildChoice == string(DEVCONTAINER)
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
