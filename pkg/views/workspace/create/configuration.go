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

var configurationHelpLine = lipgloss.NewStyle().Foreground(views.Gray).Render("enter: next  f10: advanced configuration")

type ProjectConfigurationData struct {
	Image             string
	User              string
	PostStartCommands []string
	EnvVars           map[string]string
}

func ConfigureProjects(projectList []apiclient.CreateWorkspaceRequestProject) ([]apiclient.CreateWorkspaceRequestProject, error) {
	var currentProject *apiclient.CreateWorkspaceRequestProject

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

	projectConfigurationData := &ProjectConfigurationData{
		Image:             *currentProject.Image,
		User:              *currentProject.User,
		PostStartCommands: currentProject.PostStartCommands,
		EnvVars:           *currentProject.EnvVars,
	}

	form := huh.NewForm(
		GetProjectConfigurationGroup(projectConfigurationData),
	).WithTheme(views.GetCustomTheme())

	keyMap := huh.NewDefaultKeyMap()
	keyMap.Text = huh.TextKeyMap{
		NewLine: key.NewBinding(key.WithKeys("alt+enter"), key.WithHelp("alt+enter", "new line")),
		Next:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next")),
		Prev:    key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev")),
	}

	form = form.WithKeyMap(keyMap)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	for i := range projectList {
		if projectList[i].Name == currentProject.Name {
			projectList[i].Image = &projectConfigurationData.Image
			projectList[i].User = &projectConfigurationData.User
			projectList[i].PostStartCommands = projectConfigurationData.PostStartCommands
			projectList[i].EnvVars = &projectConfigurationData.EnvVars
		}
	}

	if len(projectList) == 1 {
		return projectList, nil
	}

	return ConfigureProjects(projectList)
}

func GetProjectConfigurationGroup(projectConfiguration *ProjectConfigurationData) *huh.Group {
	group := huh.NewGroup(
		huh.NewInput().
			Title("Custom container image").
			Value(&projectConfiguration.Image),
		huh.NewInput().
			Title("Container user").
			Value(&projectConfiguration.User),
		configure.GetPostStartCommandsInput(&projectConfiguration.PostStartCommands, "Post start commands"),
		views.GetEnvVarsInput(&projectConfiguration.EnvVars),
	)

	return group
}
