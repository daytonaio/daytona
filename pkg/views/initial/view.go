// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/views"
)

const maxWidth = 80

var sigilParts = []string{
	"                  -#####=          ",
	"                 -######-           ",
	"       +###=   -######:             ",
	"       ####* -#####%-.............            ",
	"       ####*######:=##############-           ",
	"       ####* =%#-  =##############- ",
	" :*%=  ####*        ....:*#:......            ",
	"=####%==+++-           +####*.             ",
	" :*####%=               =%####+.              ",
	"   .*####%=               =%####+.         ",
	"     .*###*.          .####-=%####*.         ",
	"  :::::=%+:::::    .  .####:  =%##%-          ",
	"  #############  .*#%-.####:    +-            ",
	"  %%%%%%%%%%%%%.*####%=####:                   ",
	"             .*####%= .####:                    ",
	"           .*####%=   .####:                    ",
	"           +%##%=      ****:"}

var gradientColors = []lipgloss.AdaptiveColor{
	{Light: "#000", Dark: "#fff"},
	{Light: "#000", Dark: "#fff"},
	{Light: "#000", Dark: "#fff"},
	{Light: "#000", Dark: "#fff"},
	{Light: "#000", Dark: "#fff"},
	{Light: "#000", Dark: "#fff"},
	{Light: "#000", Dark: "#fff"},
	{Light: "#000", Dark: "#fff"},
	{Light: "#000", Dark: "#fff"},
	{Light: "#B9B9B9", Dark: "#B9B9B9"},
	{Light: "#B2B2B2", Dark: "#B2B2B2"},
	{Light: "#A4A4A4", Dark: "#A4A4A4"},
	{Light: "#969696", Dark: "#969696"},
	{Light: "#585858", Dark: "#585858"},
	{Light: "#363636", Dark: "#363636"},
	{Light: "#343434", Dark: "#343434"},
	{Light: "#333333", Dark: "#333333"},
}

var gradientSigil string

type CommandView struct {
	Command string
	Name    string
	Desc    string
}

var commandViews []CommandView = []CommandView{
	{Command: "list", Name: "daytona list", Desc: "(list all workspaces)"},
	{Command: "profile add", Name: "daytona profile add", Desc: "(run on client machine)"},
	{Command: "server api-key new", Name: "daytona server api-key new", Desc: "(create API key on this machine)"},
	{Command: "create", Name: "daytona create", Desc: "(create new workspace)\n"},
	{Command: "help", Name: "view all commands", Desc: ""},
}

type Model struct {
	form          *huh.Form
	width         int
	leftContainer string
	choice        string
}

func NewModel() Model {
	m := Model{width: maxWidth}

	var options []huh.Option[string]

	for _, commandView := range commandViews {
		options = append(options, huh.Option[string]{Key: fmt.Sprintf("%s %s", commandView.Name, lipgloss.NewStyle().Foreground(views.LightGray).Render(commandView.Desc)), Value: commandView.Command})
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("command").
				Options(options...).
				Value(&m.choice),
		),
	).WithTheme(views.GetInitialCommandTheme()).WithShowHelp(false)

	m.form = form

	return m
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = maxWidth
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	var cmds []tea.Cmd

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		command := m.form.GetString("command")
		m.choice = command
		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.form.State {
	// case huh.StateCompleted:
	// 	command := m.form.GetString("command")
	// 	m.choice = command
	// 	return command
	default:

		// Form (left side)
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		formHeader := "\n"
		formHeader += lipgloss.NewStyle().Foreground(views.Green).Render("Daytona")
		formHeader += "\n"
		formHeader += "The future of dev environemnets\n\n"
		formHeader += internal.Version + "\n\n"
		formHeader += "-------------------------------------------------------------\n\n"
		formHeader += "Get started\n"
		form := v

		body := lipgloss.JoinHorizontal(lipgloss.Top, gradientSigil, formHeader+form)

		return lipgloss.NewStyle().Padding(2, 0, 2, 2).Render(body)
	}
}

func GetCommand() (string, error) {
	for _, line := range sigilParts {
		gradientSigil += lipgloss.NewStyle().Foreground(gradientColors[0]).Render(line) + "\n"
		gradientColors = gradientColors[1:]
	}

	m := NewModel()
	p, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}

	if m, ok := p.(Model); ok && m.choice != "" {
		return m.choice, nil
	}

	return "", errors.New("no command selected")
}
