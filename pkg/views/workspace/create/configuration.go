// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
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
	EnvVars              map[string]string
}

func NewConfigurationData(buildChoice BuildChoice, devContainerFilePath string, currentProject *apiclient.CreateProjectDTO, defaults *ProjectDefaults) *ProjectConfigurationData {
	projectConfigurationData := &ProjectConfigurationData{
		BuildChoice:          string(buildChoice),
		DevcontainerFilePath: defaults.DevcontainerFilePath,
		Image:                *defaults.Image,
		User:                 *defaults.ImageUser,
		EnvVars:              map[string]string{},
	}

	if currentProject.NewConfig.Image != nil {
		projectConfigurationData.Image = *currentProject.NewConfig.Image
	}

	if currentProject.NewConfig.User != nil {
		projectConfigurationData.User = *currentProject.NewConfig.User
	}

	if currentProject.NewConfig.EnvVars != nil {
		projectConfigurationData.EnvVars = *currentProject.NewConfig.EnvVars
	}

	return projectConfigurationData
}

func ConfigureProjects(projectList *[]apiclient.CreateProjectDTO, defaults ProjectDefaults) (bool, error) {
	var currentProject *apiclient.CreateProjectDTO

	if len(*projectList) > 1 {
		// TODO: disable editing of request's existing project config entries
		currentProject = selection.GetProjectRequestFromPrompt(projectList)
		if currentProject == nil {
			return false, common.ErrCtrlCAbort
		}
	} else {
		currentProject = &((*projectList)[0])
	}

	if currentProject.NewConfig.Name == selection.DoneConfiguring.NewConfig.Name {
		return false, nil
	}

	devContainerFilePath := defaults.DevcontainerFilePath
	builderChoice := AUTOMATIC

	if currentProject.NewConfig.Build != nil {
		if currentProject.NewConfig.Build.Devcontainer != nil {
			builderChoice = DEVCONTAINER
			devContainerFilePath = *currentProject.NewConfig.Build.Devcontainer.FilePath
		}
	} else {
		if currentProject.NewConfig.Image == nil && currentProject.NewConfig.User == nil ||
			*currentProject.NewConfig.Image == *defaults.Image && *currentProject.NewConfig.User == *defaults.ImageUser {
			builderChoice = NONE
		} else {
			builderChoice = CUSTOMIMAGE
		}
	}

	projectConfigurationData := NewConfigurationData(builderChoice, devContainerFilePath, currentProject, &defaults)

	form := GetProjectConfigurationForm(projectConfigurationData)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	for i := range *projectList {
		if (*projectList)[i].NewConfig.Name == currentProject.NewConfig.Name {
			if projectConfigurationData.BuildChoice == string(NONE) {
				(*projectList)[i].NewConfig.Build = nil
				(*projectList)[i].NewConfig.Image = defaults.Image
				(*projectList)[i].NewConfig.User = defaults.ImageUser
			}

			if projectConfigurationData.BuildChoice == string(CUSTOMIMAGE) {
				(*projectList)[i].NewConfig.Build = nil
				(*projectList)[i].NewConfig.Image = &projectConfigurationData.Image
				(*projectList)[i].NewConfig.User = &projectConfigurationData.User
			}

			if projectConfigurationData.BuildChoice == string(AUTOMATIC) {
				(*projectList)[i].NewConfig.Build = &apiclient.ProjectBuildConfig{}
				(*projectList)[i].NewConfig.Image = defaults.Image
				(*projectList)[i].NewConfig.User = defaults.ImageUser
			}

			if projectConfigurationData.BuildChoice == string(DEVCONTAINER) {
				(*projectList)[i].NewConfig.Build = &apiclient.ProjectBuildConfig{
					Devcontainer: &apiclient.DevcontainerConfig{
						FilePath: &projectConfigurationData.DevcontainerFilePath,
					},
				}
				(*projectList)[i].NewConfig.Image = nil
				(*projectList)[i].NewConfig.User = nil
			}

			(*projectList)[i].NewConfig.EnvVars = &projectConfigurationData.EnvVars
		}
	}

	if len(*projectList) == 1 {
		return true, nil
	}

	return ConfigureProjects(projectList, defaults)
}

func validateDevcontainerFilename(filename string) error {
	baseName := filepath.Base(filename)
	if baseName != "devcontainer.json" && baseName != ".devcontainer.json" {
		return errors.New("filename must be devcontainer.json or .devcontainer.json")
	}
	return nil
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
				Value(&projectConfiguration.DevcontainerFilePath).Validate(validateDevcontainerFilename),
		).WithHideFunc(func() bool {
			return projectConfiguration.BuildChoice != string(DEVCONTAINER)
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
