// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"fmt"
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
	{Light: "#000", Dark: "#fff"},
	{Light: "#000", Dark: "#fff"},
	{Light: "#000", Dark: "#fff"},
	{Light: "#4e4f4f", Dark: "#B9B9B9"},
	{Light: "#686969", Dark: "#B2B2B2"},
	{Light: "#686969", Dark: "#A4A4A4"},
	{Light: "#a3a3a3", Dark: "#969696"},
	{Light: "#a3a3a3", Dark: "#585858"},
}

var gradientSigil string

type CommandView struct {
	Command string
	Name    string
	Desc    string
}

var commandViews []CommandView = []CommandView{
	{Command: "server", Name: "daytona server", Desc: "(start the Daytona Server daemon)"},
	{Command: "create", Name: "daytona create", Desc: "(create a new workspace)"},
	{Command: "code", Name: "daytona code", Desc: "(open a workspace in your preferred IDE)"},
	{Command: "git-provider add", Name: "daytona git-provider add", Desc: "(register a Git provider account)"},
	{Command: "target set", Name: "daytona target set", Desc: "(run workspaces on a remote machine)"},
	{Command: "docs", Name: "daytona docs", Desc: "(open Daytona docs in default browser)\n"},
	{Command: "help", Name: "view all commands", Desc: ""},
}

type Model struct {
	form   *huh.Form
	width  int
	choice string
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
		case "ctrl+c":
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
	case huh.StateCompleted:
		return ""
	default:
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		formHeader := "\n"
		formHeader += lipgloss.NewStyle().Foreground(views.Green).Render("Daytona")
		formHeader += "\n"
		formHeader += "The future of dev environments\n\n"
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
		return "", err
	}

	if m, ok := p.(Model); ok && m.choice != "" {
		return m.choice, nil
	}

	// return on ctrl+c
	return "", nil
}
