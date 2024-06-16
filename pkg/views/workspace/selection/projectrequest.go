// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var DoneConfiguring = apiclient.CreateWorkspaceRequestProject{Name: "DoneConfiguringName"}

type projectRequestItem struct {
	item[apiclient.CreateWorkspaceRequestProject]
	name, image, user, postStartCommands, devcontainerConfig string
	project                                                  apiclient.CreateWorkspaceRequestProject
}

type projectRequestItemDelegate struct {
	ItemDelegate[apiclient.CreateWorkspaceRequestProject]
}
type projectRequestModel struct {
	model[apiclient.CreateWorkspaceRequestProject]
}

func selectProjectRequestPrompt(projects *[]apiclient.CreateWorkspaceRequestProject, choiceChan chan<- *apiclient.CreateWorkspaceRequestProject) {
	items := []list.Item{}

	for _, project := range *projects {
		var name string
		var image string
		var user string
		var postStartCommands string
		var devcontainerConfig string

		if project.Name != "" {
			name = fmt.Sprintf("%s %s", "Project:", project.Name)
		}
		if project.Image != nil {
			image = fmt.Sprintf("%s %s", "Image:", *project.Image)
		}
		if project.User != nil {
			user = fmt.Sprintf("%s %s", "User:", *project.User)
		}
		if project.Build != nil && project.Build.Devcontainer != nil && project.Build.Devcontainer.DevContainerFilePath != nil {
			devcontainerConfig = fmt.Sprintf("%s %s", "Devcontainer Config:", *project.Build.Devcontainer.DevContainerFilePath)
		}

		newItem := projectRequestItem{name: name, image: image, user: user, project: project, devcontainerConfig: devcontainerConfig}

		newItem.SetId(name)

		if len(project.PostStartCommands) > 0 {
			postStartCommands = fmt.Sprintf("%d post start command%s", len(project.PostStartCommands), func() string {
				if len(project.PostStartCommands) == 1 {
					return ""
				} else {
					return "s"
				}
			}())
		}

		newItem.postStartCommands = postStartCommands

		items = append(items, newItem)
	}

	newItem := projectRequestItem{name: "Done configuring", image: "Return to summary view", user: "", postStartCommands: "", project: DoneConfiguring}

	items = append(items, newItem)

	l := views.GetStyledSelectList(items)
	l.SetDelegate(projectRequestItemDelegate{})

	m := projectRequestModel{}
	m.list = l
	m.list.Title = "Choose a Project To Configure"

	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("f10"),
				key.WithHelp("f10", "return to summary"),
			),
		}
	}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(projectRequestModel); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}

func GetProjectRequestFromPrompt(projects *[]apiclient.CreateWorkspaceRequestProject) *apiclient.CreateWorkspaceRequestProject {
	choiceChan := make(chan *apiclient.CreateWorkspaceRequestProject)

	go selectProjectRequestPrompt(projects, choiceChan)

	return <-choiceChan
}

func (m projectRequestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(projectRequestItem)
			if ok {
				m.choice = &i.project
			}
			return m, tea.Quit
		case "f10":
			m.choice = &DoneConfiguring
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := views.DocStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (d projectRequestItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, _ := listItem.(projectRequestItem)
	s := strings.Builder{}

	var isSelected = index == m.Index()

	baseStyles := lipgloss.NewStyle().Padding(0, 0, 0, 2)

	name := baseStyles.Copy().Render(i.Name())
	imageLine := baseStyles.Copy().Render(i.Image())
	devcontainerConfigLine := baseStyles.Copy().Render(i.DevcontainerConfig())
	userLine := baseStyles.Copy().Foreground(views.Gray).Render(i.User())
	postStartCommandsLine := baseStyles.Copy().Foreground(views.Gray).Render(i.PostStartCommands())

	// Adjust styles as the user moves through the menu
	if isSelected {
		name = selectedStyles.Copy().Foreground(views.Green).Render(i.Name())
		devcontainerConfigLine = selectedStyles.Copy().Foreground(views.DimmedGreen).Render(i.DevcontainerConfig())
		imageLine = selectedStyles.Copy().Foreground(views.DimmedGreen).Render(i.Image())
		userLine = selectedStyles.Copy().Foreground(views.Gray).Render(i.User())
		postStartCommandsLine = selectedStyles.Copy().Foreground(views.Gray).Render(i.PostStartCommands())
	}

	// Render to the terminal
	if i.project.Name == DoneConfiguring.Name {
		s.WriteRune('\n')
		s.WriteString(name)
		s.WriteRune('\n')
		s.WriteString(imageLine)
		s.WriteRune('\n')
		s.WriteRune('\n')
		s.WriteRune('\n')
	} else {
		s.WriteString(name)
		s.WriteRune('\n')
		if i.DevcontainerConfig() != "" {
			s.WriteString(devcontainerConfigLine)
		} else {
			s.WriteString(imageLine)
		}
		s.WriteRune('\n')
		s.WriteString(userLine)
		s.WriteRune('\n')
		s.WriteString(postStartCommandsLine)
		s.WriteRune('\n')
	}

	fmt.Fprint(w, s.String())
}

func (d projectRequestItemDelegate) Height() int {
	height := lipgloss.NewStyle().GetVerticalFrameSize() + 10
	return height
}

func (i projectRequestItem) Name() string               { return i.name }
func (i projectRequestItem) Image() string              { return i.image }
func (i projectRequestItem) User() string               { return i.user }
func (i projectRequestItem) PostStartCommands() string  { return i.postStartCommands }
func (i projectRequestItem) DevcontainerConfig() string { return i.devcontainerConfig }
func (i projectRequestItem) SetId(id string)            { i.id = id }
