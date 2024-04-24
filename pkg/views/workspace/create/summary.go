// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
)

type SummaryModel struct {
	lg            *lipgloss.Renderer
	styles        *Styles
	form          *huh.Form
	width         int
	quitting      bool
	workspaceName string
	projectList   []serverapiclient.CreateWorkspaceRequestProject
}

func DisplaySummaryView(workspaceName string, projectList []serverapiclient.CreateWorkspaceRequestProject, confirmCheck *bool) {
	m := NewSummaryModel(workspaceName, projectList, confirmCheck)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func RenderSummary(workspaceName string, projectList []serverapiclient.CreateWorkspaceRequestProject) string {

	output := view_util.GetStyledMainTitle(fmt.Sprintf("SUMMARY - Workspace %s", workspaceName))

	for _, project := range projectList {
		if project.Source == nil || project.Source.Repository == nil || project.Source.Repository.Url == nil {
			log.Fatal("Repository is required")
		}
	}

	output += fmt.Sprintf("\n\n%s - %s\n", lipgloss.NewStyle().Foreground(views.Blue).Render("Primary Project"), *projectList[0].Source.Repository.Url)

	// Remove the primary project from the list
	projectList = projectList[1:]

	if len(projectList) > 1 {
		output += "\n"
	}

	for i := range projectList {
		output += fmt.Sprintf("%s - %s", lipgloss.NewStyle().Foreground(views.Blue).Render(fmt.Sprintf("#%d %s", i+1, "Secondary Project")), (*projectList[i].Source.Repository.Url))
		if i < len(projectList)-1 {
			output += "\n"
		}
	}

	return output
}

func NewSummaryModel(workspaceName string, projectList []serverapiclient.CreateWorkspaceRequestProject, confirmCheck *bool) SummaryModel {
	m := SummaryModel{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.workspaceName = workspaceName
	m.projectList = projectList

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Good to go?").
				Negative("Abort").
				Value(confirmCheck),
		),
	).WithTheme(views.GetCustomTheme())

	return m
}

func (m SummaryModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m SummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd

	// Process the form
	activeForm, cmd := m.form.Update(msg)
	if f, ok := activeForm.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		// Quit when the form is done.
		m.quitting = true
		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m SummaryModel) View() string {
	if m.quitting {
		return ""
	}

	return view_util.GetBorderedMessage(RenderSummary(m.workspaceName, m.projectList)) + "\n" + m.form.View()
}
