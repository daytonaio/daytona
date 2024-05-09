// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	configure "github.com/daytonaio/daytona/pkg/views/server"
	"github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

var configurationHelpLine = lipgloss.NewStyle().Foreground(views.Gray).Render("enter: next  f10: advanced configuration")

func ConfigureProjects(projectList []serverapiclient.CreateWorkspaceRequestProject, defaultContainerImage string, defaultContainerUser string, defaultPostStartCommands string) ([]serverapiclient.CreateWorkspaceRequestProject, error) {
	var currentProject *serverapiclient.CreateWorkspaceRequestProject
	containerImage := defaultContainerImage
	containerUser := defaultContainerUser
	postStartCommands := defaultPostStartCommands

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
	if currentProject.Image != nil {
		containerImage = *currentProject.Image
	}
	if currentProject.User != nil {
		containerUser = *currentProject.User
	}
	if currentProject.PostStartCommands != nil {
		postStartCommands = util.GetJoinedCommands(currentProject.PostStartCommands)
	}

	GetProjectConfigurationGroup(&containerImage, &containerUser, &postStartCommands)

	form := huh.NewForm(
		GetProjectConfigurationGroup(&containerImage, &containerUser, &postStartCommands),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	for i := range projectList {
		if projectList[i].Name == currentProject.Name {
			projectList[i].Image = &containerImage
			projectList[i].User = &containerUser
			projectList[i].PostStartCommands = util.GetSplitCommands(postStartCommands)
		}
	}

	if len(projectList) == 1 {
		return projectList, nil
	}

	return ConfigureProjects(projectList, defaultContainerImage, defaultContainerUser, defaultPostStartCommands)
}

func GetProjectConfigurationGroup(image *string, user *string, postStartCommands *string) *huh.Group {
	group := huh.NewGroup(
		huh.NewInput().
			Title("Custom container image").
			Value(image),
		huh.NewInput().
			Title("Container user").
			Value(user),
		huh.NewInput().
			Title("Post start commands").
			Description(configure.CommandsInputHelp).
			Value(postStartCommands),
	)

	return group
}
