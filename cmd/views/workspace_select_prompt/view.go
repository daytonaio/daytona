// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_select_prompt

import (
	"fmt"
	"os"
	"strings"

	"github.com/daytonaio/daytona/cmd/views"
	workspace_proto "github.com/daytonaio/daytona/grpc/proto"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list   list.Model
	choice string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				fmt.Println("i.title")
				fmt.Println(i.title)
				m.choice = string(i.title)
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return m.choice
	}
	return docStyle.Render(m.list.View())
}

func SelectWorkspacePrompt(workspaces []*workspace_proto.WorkspaceInfoResponse, actionVerb string, choiceChan chan<- string) {

	// Initialize an empty list of items.
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, workspace := range workspaces {
		var projectNames []string
		for _, project := range workspace.Projects {
			projectNames = append(projectNames, project.Name)
		}
		newItem := item{title: workspace.Name, desc: strings.Join(projectNames, ", ")}
		items = append(items, newItem)
	}

	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(views.Blue).
		Foreground(views.Blue).
		Bold(true).
		Padding(0, 0, 0, 1)

	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Copy().Foreground(views.DimmedBlue)

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	m := model{list: l}

	m.list.Title = "SELECT A WORKSPACE TO " + strings.ToUpper(actionVerb)
	m.list.Styles.Title = lipgloss.NewStyle().Foreground(views.Green).Bold(true)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model); ok && m.choice != "" {
		choiceChan <- m.choice
	} else {
		choiceChan <- ""
	}
}

func GetWorkspaceNameFromPrompt(workspaces []*workspace_proto.WorkspaceInfoResponse, actionVerb string) string {
	choseWorkspaceName := ""
	choiceChan := make(chan string)

	go SelectWorkspacePrompt(workspaces, actionVerb, choiceChan)

	choseWorkspaceName = <-choiceChan

	return choseWorkspaceName
}
