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

func ConfigureProjects(projectList []serverapiclient.CreateWorkspaceRequestProject) ([]serverapiclient.CreateWorkspaceRequestProject, error) {
	var containerImage string
	var osUser string
	var doneCheck bool

	projectName := selection.GetProjectRequestFromPrompt(projectList)
	if projectName == "" {
		return projectList, errors.New("project ID is required")
	}

	GetProjectConfigurationGroup(&containerImage, &osUser)

	form := huh.NewForm(
		GetProjectConfigurationGroup(&containerImage, &osUser),
		GetDoneCheckGroup(&doneCheck),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	for i := range projectList {
		if projectList[i].Name == projectName {
			projectList[i].Image = &containerImage
			projectList[i].User = &osUser
		}
	}

	if !doneCheck {
		projectList, err = ConfigureProjects(projectList)
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
