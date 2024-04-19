// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package started

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
	getting_started "github.com/daytonaio/daytona/pkg/views/initial"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"golang.org/x/term"
)

const propertyNameWidth = 16
const minTUIWidth = 80
const maxTUIWidth = 140

// func RenderOld(apiPort uint32, frpcUrl string) {
// 	output := "\n"
// 	output += view_util.GetStyledMainTitle("Daytona") + "\n\n"

// 	output += fmt.Sprintf("## Daytona Server running on port: %d.\n\n", apiPort)

// 	output += view_util.GetSeparatorString() + "\n\n"

// 	output += "You can now begin developing locally\n\n"

// 	output += view_util.GetSeparatorString() + "\n\n"

// 	output += fmt.Sprintf("If you want to connect to the server remotely:\n\n1. Create an API key on this machine:\ndaytona server api-key new\n\n2. On the client machine run:\ndaytona profile add -a %s -k API_KEY", frpcUrl)

// 	var width int
// 	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
// 	if err != nil || terminalWidth < minTUIWidth {
// 		fmt.Println(output)
// 		return
// 	}
// 	width = terminalWidth - 20
// 	if width > maxTUIWidth {
// 		width = maxTUIWidth
// 	}

// 	renderTUIView(output, width)
// }

// func renderTUIView(output string, terminalWidth int) {
// 	output = lipgloss.NewStyle().PaddingLeft(3).Render(output)

// 	fmt.Println(lipgloss.
// 		NewStyle().
// 		BorderForeground(views.LightGray).
// 		Border(lipgloss.RoundedBorder()).Width(terminalWidth).
// 		Render(output),
// 	)
// }

//

type Model struct {
	form          *huh.Form
	width         int
	leftContainer string
	apiPort       uint32
	frpcUrl       string
	isDaemonMode  bool
}

func NewModel(apiPort uint32, frpcUrl string, isDaemonMode bool) Model {
	m := Model{width: maxTUIWidth, apiPort: apiPort, frpcUrl: frpcUrl, isDaemonMode: isDaemonMode}

	var options []huh.Option[string]

	options = append(options, huh.Option[string]{Key: fmt.Sprintf("%s %s", "test", "test"), Value: "test"})

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("command").
				Options(options...),
		),
	).WithTheme(views.GetInitialCommandTheme())

	m.form = form

	return m
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = maxTUIWidth
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			fmt.Println()
			_, err := getting_started.GetCommand()
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	return m, cmd
}

func (m Model) View() string {
	var width int

	switch m.form.State {
	case huh.StateCompleted:
		return "done"
	default:
		output := view_util.GetStyledMainTitle("Daytona") + "\n\n"
		output += fmt.Sprintf("## Daytona Server is running on port: %d\n\n", m.apiPort)
		output += view_util.GetSeparatorString() + "\n\n"
		output += "You may now begin developing locally"
		if m.isDaemonMode {
			output += ". Press Enter to get started"
		}
		output += "\n\n"

		// output += fmt.Sprintf("If you want to connect to the server remotely:\n\n")

		// output += fmt.Sprintf("1. Create an API key on this machine: ")
		// output += lipgloss.NewStyle().Foreground(views.Green).Render("daytona server api-key new") + "\n"
		// output += fmt.Sprintf("2. Add a profile on the client machine: \n\t")
		// output += lipgloss.NewStyle().Foreground(views.Green).Render(fmt.Sprintf("daytona profile add -a %s -k API_KEY", m.frpcUrl)) + "\n\n"

		// output += view_util.GetSeparatorString() + "\n\n"

		// output += "Press Enter to create an API key and copy the complete client command to clipboard automatically"

		terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil || terminalWidth < minTUIWidth {
			fmt.Println(output)
			return ""
		}
		width = terminalWidth - 20
		if width > maxTUIWidth {
			width = maxTUIWidth
		}

		output = lipgloss.NewStyle().PaddingLeft(3).Render(output)

		return "\n" + lipgloss.
			NewStyle().
			BorderForeground(views.LightGray).
			Border(lipgloss.RoundedBorder()).Width(width).
			Render(output) + "\n"
	}
}

func Render(apiPort uint32, frpcUrl string, isDaemonMode bool) {
	_, err := tea.NewProgram(NewModel(apiPort, frpcUrl, isDaemonMode)).Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
