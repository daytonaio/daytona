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
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

const (
	DEVCONTAINER_FILEPATH = ".devcontainer/devcontainer.json"
)

var configurationHelpLine = lipgloss.NewStyle().Foreground(views.Gray).Render("enter: next  f10: configuration screen")

type WorkspaceConfigurationData struct {
	Name                 string
	BuildChoice          string
	DevcontainerFilePath string
	Image                string
	User                 string
	EnvVars              map[string]string
}

func NewConfigurationData(buildChoice views_util.BuildChoice, devContainerFilePath string, currentWorkspace *apiclient.CreateWorkspaceDTO, defaults *views_util.WorkspaceConfigDefaults) *WorkspaceConfigurationData {
	workspaceConfigurationData := &WorkspaceConfigurationData{
		Name:                 currentWorkspace.Name,
		BuildChoice:          string(buildChoice),
		DevcontainerFilePath: defaults.DevcontainerFilePath,
		Image:                *defaults.Image,
		User:                 *defaults.ImageUser,
		EnvVars:              map[string]string{},
	}

	if currentWorkspace.Image != nil {
		workspaceConfigurationData.Image = *currentWorkspace.Image
	}

	if currentWorkspace.User != nil {
		workspaceConfigurationData.User = *currentWorkspace.User
	}

	if currentWorkspace.EnvVars != nil {
		workspaceConfigurationData.EnvVars = currentWorkspace.EnvVars
	}

	return workspaceConfigurationData
}

func RunWorkspaceConfiguration(workspaceList *[]apiclient.CreateWorkspaceDTO, defaults views_util.WorkspaceConfigDefaults) (bool, error) {
	var currentWorkspace *apiclient.CreateWorkspaceDTO

	if len(*workspaceList) > 1 {
		currentWorkspace = selection.GetWorkspaceRequestFromPrompt(workspaceList)
		if currentWorkspace == nil {
			return false, common.ErrCtrlCAbort
		}
	} else {
		currentWorkspace = &((*workspaceList)[0])
	}

	if currentWorkspace.Name == selection.DoneConfiguring.Name {
		return false, nil
	}

	devContainerFilePath := defaults.DevcontainerFilePath
	builderChoice := views_util.AUTOMATIC

	if currentWorkspace.BuildConfig != nil {
		if currentWorkspace.BuildConfig.Devcontainer != nil {
			builderChoice = views_util.DEVCONTAINER
			devContainerFilePath = currentWorkspace.BuildConfig.Devcontainer.FilePath
		}
	} else {
		if currentWorkspace.Image == nil && currentWorkspace.User == nil ||
			*currentWorkspace.Image == *defaults.Image && *currentWorkspace.User == *defaults.ImageUser {
			builderChoice = views_util.NONE
		} else {
			builderChoice = views_util.CUSTOMIMAGE
		}
	}

	workspaceConfigurationData := NewConfigurationData(builderChoice, devContainerFilePath, currentWorkspace, &defaults)

	form := GetWorkspaceConfigurationForm(workspaceConfigurationData)
	err := form.WithProgramOptions(tea.WithAltScreen()).Run()
	if err != nil {
		log.Fatal(err)
	}

	for i := range *workspaceList {
		if (*workspaceList)[i].Id != currentWorkspace.Id {
			continue
		}

		if workspaceConfigurationData.BuildChoice == string(views_util.NONE) {
			(*workspaceList)[i].BuildConfig = nil
			(*workspaceList)[i].Image = defaults.Image
			(*workspaceList)[i].User = defaults.ImageUser
		}

		if workspaceConfigurationData.BuildChoice == string(views_util.CUSTOMIMAGE) {
			(*workspaceList)[i].BuildConfig = nil
			(*workspaceList)[i].Image = &workspaceConfigurationData.Image
			(*workspaceList)[i].User = &workspaceConfigurationData.User
		}

		if workspaceConfigurationData.BuildChoice == string(views_util.AUTOMATIC) {
			(*workspaceList)[i].BuildConfig = &apiclient.BuildConfig{}
			(*workspaceList)[i].Image = defaults.Image
			(*workspaceList)[i].User = defaults.ImageUser
		}

		if workspaceConfigurationData.BuildChoice == string(views_util.DEVCONTAINER) {
			(*workspaceList)[i].BuildConfig = &apiclient.BuildConfig{
				Devcontainer: &apiclient.DevcontainerConfig{
					FilePath: workspaceConfigurationData.DevcontainerFilePath,
				},
			}
			(*workspaceList)[i].Image = nil
			(*workspaceList)[i].User = nil
		}

		(*workspaceList)[i].Name = workspaceConfigurationData.Name
		(*workspaceList)[i].EnvVars = workspaceConfigurationData.EnvVars
	}

	if len(*workspaceList) == 1 {
		return true, nil
	}

	return RunWorkspaceConfiguration(workspaceList, defaults)
}

func validateDevcontainerFilename(filename string) error {
	baseName := filepath.Base(filename)
	if baseName != "devcontainer.json" && baseName != ".devcontainer.json" {
		return errors.New("filename must be devcontainer.json or .devcontainer.json")
	}
	return nil
}

func GetWorkspaceConfigurationForm(workspaceConfiguration *WorkspaceConfigurationData) *huh.Form {
	buildOptions := []huh.Option[string]{
		{Key: "Automatic", Value: string(views_util.AUTOMATIC)},
		{Key: "Devcontainer", Value: string(views_util.DEVCONTAINER)},
		{Key: "Custom image", Value: string(views_util.CUSTOMIMAGE)},
		{Key: "None", Value: string(views_util.NONE)},
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Value(&workspaceConfiguration.Name),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a build configuration").
				Options(
					buildOptions...,
				).
				Value(&workspaceConfiguration.BuildChoice),
		).WithHeight(8),
		huh.NewGroup(
			huh.NewInput().
				Title("Custom container image").
				Value(&workspaceConfiguration.Image),
			huh.NewInput().
				Title("Container user").
				Value(&workspaceConfiguration.User),
		).WithHeight(5).WithHideFunc(func() bool {
			return workspaceConfiguration.BuildChoice != string(views_util.CUSTOMIMAGE)
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Devcontainer file path").
				Value(&workspaceConfiguration.DevcontainerFilePath).Validate(validateDevcontainerFilename),
		).WithHeight(5).WithHideFunc(func() bool {
			return workspaceConfiguration.BuildChoice != string(views_util.DEVCONTAINER)
		}),
		huh.NewGroup(
			views.GetEnvVarsInput(&workspaceConfiguration.EnvVars),
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
