// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package initialize

import (
	"os"
	"sort"

	"github.com/daytonaio/daytona/pkg/server/event_bus"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

type EventMsg struct {
	Event   string
	Payload string
}

type ClearScreenMsg struct{}

var colors = views.ColorGrid(5, 5)

var workspaceStatusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFF7DB")).
	Background(lipgloss.Color(colors[4][0])).
	Bold(true).
	PaddingLeft(2).
	PaddingRight(2)

var projectViewStyle = lipgloss.NewStyle().
	Border(lipgloss.HiddenBorder()).
	BorderForeground(lipgloss.Color(colors[4][2])).
	Foreground(lipgloss.Color("#FFF7DB")).
	Margin(1).
	MarginBottom(0).
	Padding(1).
	PaddingLeft(2).
	PaddingRight(2)

var projectNameStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(views.Blue).
	PaddingLeft(2)

var projectStatusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(colors[0][4])).
	PaddingLeft(2)

type InitWorkspaceViewProjectExtensionModel struct {
	Name  string
	State string
	Info  string
}

type InitWorkspaceViewProjectModel struct {
	Name       string
	State      string
	Extensions map[string]InitWorkspaceViewProjectExtensionModel
}

type InitWorkspaceViewModel struct {
	State    string
	Err      string
	Projects map[string]InitWorkspaceViewProjectModel
	spinner  spinner.Model
	done     bool
}

func GetInitialModel() InitWorkspaceViewModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))

	return InitWorkspaceViewModel{
		State:    "Pending",
		spinner:  s,
		done:     false,
		Projects: map[string]InitWorkspaceViewProjectModel{},
	}
}

func (m InitWorkspaceViewModel) Init() tea.Cmd {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	workspaceStatusStyle.Width(physicalWidth)
	projectViewStyle.Width(physicalWidth)
	return m.spinner.Tick
}

func (m InitWorkspaceViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
		workspaceStatusStyle.Width(physicalWidth)
		projectViewStyle.Width(physicalWidth - 4)
		return m, m.spinner.Tick
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			reallyQuit := false

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Do you want to exit?").
						Description("This will not end the workspace initialization process.").
						Value(&reallyQuit),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				return m, tea.Quit
			}

			if !reallyQuit {
				return m, nil
			}
			return m, tea.Quit
		}
		return m, nil
	case EventMsg:
		return m.HandleEvent(msg), m.spinner.Tick
	case ClearScreenMsg:
		m.done = true
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m InitWorkspaceViewModel) HandleWorkspaceInfo(msg *types.WorkspaceInfo) InitWorkspaceViewModel {
	// TODO: handle
	// for _, projectInfo := range msg.Projects {
	// 	if _, ok := m.Projects[projectInfo.Name]; !ok {
	// 		continue
	// 	}

	// 	for _, extension := range projectInfo.Extensions {
	// 		if _, ok := m.Projects[projectInfo.Name].Extensions[extension.Name]; !ok {
	// 			continue
	// 		}

	// 		m.Projects[projectInfo.Name].Extensions[extension.Name] = InitWorkspaceViewProjectExtensionModel{
	// 			Name:  extension.Name,
	// 			State: "Started",
	// 			Info:  extension.Info,
	// 		}
	// 	}
	// }
	return m
}

func (m InitWorkspaceViewModel) HandleEvent(msg EventMsg) InitWorkspaceViewModel {
	return m.handleWorkspaceEvent(msg).handleProjectEvent(msg)
}

func (m InitWorkspaceViewModel) View() string {
	if m.done {
		return ""
	}

	sortedProjects := []string{}
	for project := range m.Projects {
		sortedProjects = append(sortedProjects, project)
	}
	sort.Strings(sortedProjects)

	projects := ""
	for _, project := range sortedProjects {
		projects += projectRender(m.Projects[project])
	}

	spinner := ""
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if physicalWidth > 80 {
		physicalWidth = 80
	}

	switch m.State {
	case "Started":
		workspaceStatusStyle.Background(lipgloss.Color(colors[4][4]))
		workspaceStatusStyle.Width(physicalWidth)
	case "Starting":
		workspaceStatusStyle.Background(lipgloss.Color(colors[3][4]))
		workspaceStatusStyle.Width(physicalWidth - 2)
		spinner = m.spinner.View()
	case "Initializing projects":
		workspaceStatusStyle.Background(lipgloss.Color(colors[2][4]))
		workspaceStatusStyle.Width(physicalWidth - 2)
		spinner = m.spinner.View()
	default:
		workspaceStatusStyle.Background(lipgloss.Color(colors[0][4]))
		workspaceStatusStyle.Width(physicalWidth - 2)
		spinner = m.spinner.View()
	}

	output := "  " +
		spinner + workspaceStatusStyle.Render(m.State) +
		projects +
		"\n\n"

	return output
}

func projectRender(project InitWorkspaceViewProjectModel) string {
	projectState := ""
	switch project.State {
	case "Initialized":
		projectState = projectStatusStyle.Foreground(lipgloss.Color(colors[1][4])).Render(project.State)
	case "Starting":
		projectState = projectStatusStyle.Foreground(lipgloss.Color(colors[2][4])).Render(project.State)
	case "Started":
		projectState = projectStatusStyle.Foreground(lipgloss.Color(colors[4][4])).Render(project.State)
	default:
		projectState = projectStatusStyle.Render(project.State)
	}

	sortedExtensions := []string{}
	for extension := range project.Extensions {
		sortedExtensions = append(sortedExtensions, extension)
	}
	sort.Strings(sortedExtensions)

	extensions := [][]string{}
	for _, extensionName := range sortedExtensions {
		extension := project.Extensions[extensionName]
		extensions = append(extensions, []string{extension.Name, extension.State, extension.Info})
	}

	extensionsTable := table.New().
		Border(lipgloss.HiddenBorder()).
		Rows(extensions...)

	projectView := "Project" + projectNameStyle.Render(project.Name) + "\n" + "State  " + projectState + "\n" + extensionsTable.Render()

	switch project.State {
	case "Started":
		return projectViewStyle.Border(lipgloss.NormalBorder()).Render(projectView)
	default:
		return projectViewStyle.Border(lipgloss.HiddenBorder()).Render(projectView)
	}
}

func (m InitWorkspaceViewModel) handleWorkspaceEvent(msg EventMsg) InitWorkspaceViewModel {
	workspaceEventPayload, err := event_bus.UnmarshallWorkspaceEventPayload(msg.Payload)
	if err != nil {
		return m
	}

	switch msg.Event {
	//	workspace events
	case string(event_bus.WorkspaceEventCreatingNetwork):
		m.State = "Creating network"
	case string(event_bus.WorkspaceEventNetworkCreated):
		m.State = "Network created"
	case string(event_bus.WorkspaceEventStarting):
		m.State = "Starting"
	case string(event_bus.WorkspaceEventStarted):
		m.State = "Started"
	case string(event_bus.WorkspaceEventCreating):
		m.State = "Creating projects"
		m.Projects[workspaceEventPayload.ProjectName] = InitWorkspaceViewProjectModel{
			Name:       workspaceEventPayload.ProjectName,
			State:      "Creating",
			Extensions: map[string]InitWorkspaceViewProjectExtensionModel{},
		}
	}

	return m
}

func (m InitWorkspaceViewModel) handleProjectEvent(msg EventMsg) InitWorkspaceViewModel {
	projectEventPayload, err := event_bus.UnmarshallProjectEventPayload(msg.Payload)
	if err != nil || projectEventPayload.ProjectName == "" {
		return m
	}

	// Handle unordered project events
	if _, ok := m.Projects[projectEventPayload.ProjectName]; !ok {
		m.Projects[projectEventPayload.ProjectName] = InitWorkspaceViewProjectModel{
			Name:       projectEventPayload.ProjectName,
			State:      "Creating",
			Extensions: map[string]InitWorkspaceViewProjectExtensionModel{},
		}
	}

	newProjectState := ""
	newExtensionState := ""

	switch msg.Event {
	case string(event_bus.ProjectEventCloningRepo):
		newProjectState = "Cloning repository"
	case string(event_bus.ProjectEventRepoCloned):
		newProjectState = "Repository cloned"
	case string(event_bus.ProjectEventInitializing):
		newProjectState = "Initializing"
	case string(event_bus.ProjectEventInitialized):
		newProjectState = "Initialized"
	case string(event_bus.ProjectEventStarting):
		newProjectState = "Starting"
	case string(event_bus.ProjectEventStarted):
		newProjectState = "Started"
	case string(event_bus.ProjectEventPreparingExtension):
		newExtensionState = "Preparing"
	case string(event_bus.ProjectEventInitializingExtension):
		newExtensionState = "Initializing"
	case string(event_bus.ProjectEventStartingExtension):
		newExtensionState = "Starting"
	}

	if newProjectState != "" {
		m.Projects[projectEventPayload.ProjectName] = InitWorkspaceViewProjectModel{
			Name:       projectEventPayload.ProjectName,
			State:      newProjectState,
			Extensions: m.Projects[projectEventPayload.ProjectName].Extensions,
		}
	}

	if newExtensionState != "" {
		oldInfo := ""
		oldExtension, ok := m.Projects[projectEventPayload.ProjectName].Extensions[projectEventPayload.ExtensionName]
		if ok {
			oldInfo = oldExtension.Info
		}

		m.Projects[projectEventPayload.ProjectName].Extensions[projectEventPayload.ExtensionName] = InitWorkspaceViewProjectExtensionModel{
			Name:  projectEventPayload.ExtensionName,
			State: newExtensionState,
			Info:  oldInfo,
		}
	}

	return m
}
