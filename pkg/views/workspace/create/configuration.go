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
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

var configurationHelpLine = lipgloss.NewStyle().Foreground(views.Gray).Render("enter: next  f10: advanced configuration")

func ConfigureProjects(projectList []serverapiclient.CreateWorkspaceRequestProject, defaultContainerImage string, defaultContainerUser string, defaultPostStartCommands string) ([]serverapiclient.CreateWorkspaceRequestProject, error) {
	containerImage := defaultContainerImage
	containerUser := defaultContainerUser
	postStartCommands := defaultPostStartCommands
	var doneCheck bool

	project := selection.GetProjectRequestFromPrompt(projectList, defaultContainerUser)
	if project == nil {
		return projectList, errors.New("project is required")
	}
	if project.Image != nil {
		containerImage = *project.Image
	}
	if project.User != nil {
		containerUser = *project.User
	}
	if project.PostStartCommands != nil {
		postStartCommands = view_util.GetJoinedCommands(project.PostStartCommands)
	}

	GetProjectConfigurationGroup(&containerImage, &containerUser, &postStartCommands)

	form := huh.NewForm(
		GetProjectConfigurationGroup(&containerImage, &containerUser, &postStartCommands),
		GetDoneCheckGroup(&doneCheck),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	for i := range projectList {
		if projectList[i].Name == project.Name {
			projectList[i].Image = &containerImage
			projectList[i].User = &containerUser
			projectList[i].PostStartCommands = view_util.GetSplitCommands(postStartCommands)
		}
	}

	if !doneCheck {
		projectList, err = ConfigureProjects(projectList, defaultContainerImage, defaultContainerUser, defaultPostStartCommands)
		if err != nil {
			return projectList, err
		}
	}

	return projectList, nil
}

func GetDoneCheckGroup(doneCheck *bool) *huh.Group {
	group := huh.NewGroup(
		huh.NewConfirm().
			Title("Done configuring projects?").
			Value(doneCheck),
	)

	return group
}
