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

var doneConfiguringName = "DoneConfiguringName"
var DoneConfiguring = apiclient.CreateWorkspaceDTO{
	Name: doneConfiguringName,
}

type workspaceRequestItem struct {
	item[apiclient.CreateWorkspaceDTO]
	name, image, user, devcontainerConfig string
	workspace                             apiclient.CreateWorkspaceDTO
}

type workspaceRequestItemDelegate struct {
	ItemDelegate[apiclient.CreateWorkspaceDTO]
}
type workspaceRequestModel struct {
	model[apiclient.CreateWorkspaceDTO]
}

func selectWorkspaceRequestPrompt(workspaces *[]apiclient.CreateWorkspaceDTO, choiceChan chan<- *apiclient.CreateWorkspaceDTO) {
	items := []list.Item{}

	for _, workspace := range *workspaces {
		var name string
		var image string
		var user string
		var devcontainerConfig string

		name = fmt.Sprintf("%s %s", "Workspace:", workspace.Name)
		if workspace.Image != nil {
			image = fmt.Sprintf("%s %s", "Image:", *workspace.Image)
		}
		if workspace.User != nil {
			user = fmt.Sprintf("%s %s", "User:", *workspace.User)
		}
		if workspace.BuildConfig != nil && workspace.BuildConfig.Devcontainer != nil {
			devcontainerConfig = fmt.Sprintf("%s %s", "Devcontainer Config:", workspace.BuildConfig.Devcontainer.FilePath)
		}

		newItem := workspaceRequestItem{name: name, image: image, user: user, workspace: workspace, devcontainerConfig: devcontainerConfig}

		newItem.SetId(name)

		items = append(items, newItem)
	}

	newItem := workspaceRequestItem{name: "Done configuring", image: "Return to summary view", user: "", workspace: DoneConfiguring}

	items = append(items, newItem)

	l := views.GetStyledSelectList(items)
	l.SetDelegate(workspaceRequestItemDelegate{})

	m := workspaceRequestModel{}
	m.list = l
	m.list.Title = "Choose a Workspace To Configure"

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

	if m, ok := p.(workspaceRequestModel); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}

func GetWorkspaceRequestFromPrompt(workspaces *[]apiclient.CreateWorkspaceDTO) *apiclient.CreateWorkspaceDTO {
	choiceChan := make(chan *apiclient.CreateWorkspaceDTO)

	go selectWorkspaceRequestPrompt(workspaces, choiceChan)

	return <-choiceChan
}

func (m workspaceRequestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(workspaceRequestItem)
			if ok {
				m.choice = &i.workspace
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

func (d workspaceRequestItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, _ := listItem.(workspaceRequestItem)
	s := strings.Builder{}

	var isSelected = index == m.Index()

	baseStyles := lipgloss.NewStyle().Padding(0, 0, 0, 2)

	name := baseStyles.Render(i.Name())
	imageLine := baseStyles.Render(i.Image())
	devcontainerConfigLine := baseStyles.Render(i.DevcontainerConfig())
	userLine := baseStyles.Foreground(views.Gray).Render(i.User())

	// Adjust styles as the user moves through the menu
	if isSelected {
		name = selectedStyles.Foreground(views.Green).Render(i.Name())
		devcontainerConfigLine = selectedStyles.Foreground(views.DimmedGreen).Render(i.DevcontainerConfig())
		imageLine = selectedStyles.Foreground(views.DimmedGreen).Render(i.Image())
		userLine = selectedStyles.Foreground(views.Gray).Render(i.User())
	}

	// Render to the terminal
	if i.workspace.Name == DoneConfiguring.Name {
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
	}

	fmt.Fprint(w, s.String())
}

func (d workspaceRequestItemDelegate) Height() int {
	height := lipgloss.NewStyle().GetVerticalFrameSize() + 10
	return height
}

func (i workspaceRequestItem) Name() string               { return i.name }
func (i workspaceRequestItem) Image() string              { return i.image }
func (i workspaceRequestItem) User() string               { return i.user }
func (i workspaceRequestItem) DevcontainerConfig() string { return i.devcontainerConfig }
func (i workspaceRequestItem) SetId(id string)            { i.id = id }
