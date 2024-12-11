// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
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

func NewConfigurationData(buildChoice views_util.BuildChoice, devContainerFilePath string, currentProject *apiclient.CreateProjectDTO, defaults *views_util.ProjectConfigDefaults) *ProjectConfigurationData {
	projectConfigurationData := &ProjectConfigurationData{
		BuildChoice:          string(buildChoice),
		DevcontainerFilePath: defaults.DevcontainerFilePath,
		Image:                *defaults.Image,
		User:                 *defaults.ImageUser,
		EnvVars:              map[string]string{},
	}

	if currentProject.Image != nil {
		projectConfigurationData.Image = *currentProject.Image
	}

	if currentProject.User != nil {
		projectConfigurationData.User = *currentProject.User
	}

	if currentProject.EnvVars != nil {
		projectConfigurationData.EnvVars = currentProject.EnvVars
	}

	return projectConfigurationData
}

func RunProjectConfiguration(projectList *[]apiclient.CreateProjectDTO, defaults views_util.ProjectConfigDefaults, isConfigImport bool) (bool, error) {
	var currentProject *apiclient.CreateProjectDTO

	if len(*projectList) > 1 {
		currentProject = selection.GetProjectRequestFromPrompt(projectList)
		if currentProject == nil {
			return false, common.ErrCtrlCAbort
		}
	} else {
		currentProject = &((*projectList)[0])
	}

	if currentProject.Name == selection.DoneConfiguring.Name {
		return false, nil
	}

	devContainerFilePath := defaults.DevcontainerFilePath
	builderChoice := views_util.AUTOMATIC

	if currentProject.BuildConfig != nil {
		if currentProject.BuildConfig.Devcontainer != nil {
			builderChoice = views_util.DEVCONTAINER
			devContainerFilePath = currentProject.BuildConfig.Devcontainer.FilePath
		}
	} else {
		if currentProject.Image == nil && currentProject.User == nil ||
			*currentProject.Image == *defaults.Image && *currentProject.User == *defaults.ImageUser {
			builderChoice = views_util.NONE
		} else {
			builderChoice = views_util.CUSTOMIMAGE
		}
	}

	projectConfigurationData := NewConfigurationData(builderChoice, devContainerFilePath, currentProject, &defaults)

	if !isConfigImport {
		form := GetProjectConfigurationForm(projectConfigurationData)
		err := form.WithProgramOptions(tea.WithAltScreen()).Run()
		if err != nil {
			log.Fatal(err)
		}
	}

	for i := range *projectList {
		if (*projectList)[i].Name == currentProject.Name {
			if projectConfigurationData.BuildChoice == string(views_util.NONE) {
				(*projectList)[i].BuildConfig = nil
				(*projectList)[i].Image = defaults.Image
				(*projectList)[i].User = defaults.ImageUser
			}

			if projectConfigurationData.BuildChoice == string(views_util.CUSTOMIMAGE) {
				(*projectList)[i].BuildConfig = nil
				(*projectList)[i].Image = &projectConfigurationData.Image
				(*projectList)[i].User = &projectConfigurationData.User
			}

			if projectConfigurationData.BuildChoice == string(views_util.AUTOMATIC) {
				(*projectList)[i].BuildConfig = &apiclient.BuildConfig{}
				(*projectList)[i].Image = defaults.Image
				(*projectList)[i].User = defaults.ImageUser
			}

			if projectConfigurationData.BuildChoice == string(views_util.DEVCONTAINER) {
				(*projectList)[i].BuildConfig = &apiclient.BuildConfig{
					Devcontainer: &apiclient.DevcontainerConfig{
						FilePath: projectConfigurationData.DevcontainerFilePath,
					},
				}
				(*projectList)[i].Image = nil
				(*projectList)[i].User = nil
			}

			(*projectList)[i].EnvVars = projectConfigurationData.EnvVars
		}
	}

	if len(*projectList) == 1 {
		return true, nil
	}

	return RunProjectConfiguration(projectList, defaults, isConfigImport)
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
		{Key: "Automatic", Value: string(views_util.AUTOMATIC)},
		{Key: "Devcontainer", Value: string(views_util.DEVCONTAINER)},
		{Key: "Custom image", Value: string(views_util.CUSTOMIMAGE)},
		{Key: "None", Value: string(views_util.NONE)},
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a build configuration").
				Options(
					buildOptions...,
				).
				Value(&projectConfiguration.BuildChoice),
		).WithHeight(8),
		huh.NewGroup(
			huh.NewInput().
				Title("Custom container image").
				Value(&projectConfiguration.Image),
			huh.NewInput().
				Title("Container user").
				Value(&projectConfiguration.User),
		).WithHeight(5).WithHideFunc(func() bool {
			return projectConfiguration.BuildChoice != string(views_util.CUSTOMIMAGE)
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Devcontainer file path").
				Value(&projectConfiguration.DevcontainerFilePath).Validate(validateDevcontainerFilename),
		).WithHeight(5).WithHideFunc(func() bool {
			return projectConfiguration.BuildChoice != string(views_util.DEVCONTAINER)
		}),
		huh.NewGroup(
			views.GetEnvVarsInput(&projectConfiguration.EnvVars),
		).WithHeight(12),
	).WithTheme(views.GetCustomTheme())

	keyMap := huh.NewDefaultKeyMap()
	keyMap.Text = huh.TextKeyMap{
		NewLine: key.NewBinding(key.WithKeys("alt+enter"), key.WithHelp("alt+enter", "new line")),
		Next:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next")),
		Prev:    key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev")),
		Submit:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	}

	form = form.WithKeyMap(keyMap)

	return form
}
