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
	selection "github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

var configurationHelpLine = lipgloss.NewStyle().Foreground(views.Gray).Render("enter: next  f10: advanced configuration")

func ConfigureProjects(projectList []serverapiclient.CreateWorkspaceRequestProject, defaultContainerImage string, defaultContainerUser string) ([]serverapiclient.CreateWorkspaceRequestProject, error) {
	containerImage := defaultContainerImage
	containerUser := defaultContainerUser
	var doneCheck bool

	project := selection.GetProjectRequestFromPrompt(projectList)
	if project == nil {
		return projectList, errors.New("project is required")
	}
	if project.Image != nil {
		containerImage = *project.Image
	}
	if project.User != nil {
		containerUser = *project.User
	}

	GetProjectConfigurationGroup(&containerImage, &containerUser)

	form := huh.NewForm(
		GetProjectConfigurationGroup(&containerImage, &containerUser),
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
		}
	}

	if !doneCheck {
		projectList, err = ConfigureProjects(projectList, defaultContainerImage, defaultContainerUser)
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
