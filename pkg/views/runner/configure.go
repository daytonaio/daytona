// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/runner"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"
)

type Model struct {
	form     *huh.Form
	quitting bool
	config   *runner.Config
	keymap   keymap
	help     help.Model
}

type keymap struct {
	submit key.Binding
}

func NewModel(config *runner.Config) Model {
	m := Model{
		config: config,
		help:   help.New(),
		keymap: keymap{
			submit: key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "submit")),
		},
	}

	m.form = m.createForm()
	return m
}
func (m *Model) createForm() *huh.Form {
	logFileMaxSize := strconv.Itoa(int(m.config.LogFile.MaxSize))
	logFileMaxBackups := strconv.Itoa(int(m.config.LogFile.MaxBackups))
	logFileMaxAge := strconv.Itoa(int(m.config.LogFile.MaxAge))
	apiPort := strconv.Itoa(int(m.config.ApiPort))

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("ID").
				Description("Unique ID generated by the Daytona Server").
				Value(&m.config.Id),
			huh.NewInput().
				Title("Name").
				Description("Unique name set on the Daytona Server").
				Value(&m.config.Name),
			huh.NewInput().
				Title("Runner API Port").
				Description("Port used for exposing runner health-check endpoint").
				Value(&apiPort).
				Validate(util.CreatePortValidator(&apiPort, &m.config.ApiPort)),
			huh.NewInput().
				Title("Server API URL").
				Value(&m.config.ServerApiUrl),
			huh.NewInput().
				Title("Server API Key").
				EchoMode(huh.EchoModePassword).
				Value(&m.config.ServerApiKey),
			huh.NewInput().
				Title("Providers Directory").
				Description("Directory will be created if it does not exist").
				Value(&m.config.ProvidersDir),
			huh.NewConfirm().
				Title("Telemetry Enabled").
				Value(&m.config.TelemetryEnabled),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Log File Path").
				Description("File will be created if it does not exist").
				Value(&m.config.LogFile.Path).
				Validate(func(s string) error {
					_, err := os.Stat(s)
					if os.IsNotExist(err) {
						err = os.MkdirAll(filepath.Dir(s), 0755)
						if err != nil {
							return err
						}
						_, err = os.Create(s)
					}
					return err
				}),
			huh.NewInput().
				Title("Log File Max Size").
				Description("In megabytes").
				Value(&logFileMaxSize).
				Validate(util.CreateIntValidator(&logFileMaxSize, &m.config.LogFile.MaxSize)),
			huh.NewInput().
				Title("Log File Max Backups").
				Value(&logFileMaxBackups).
				Validate(util.CreateIntValidator(&logFileMaxBackups, &m.config.LogFile.MaxBackups)),
			huh.NewInput().
				Title("Log File Max Age").
				Description("In days").
				Value(&logFileMaxAge).
				Validate(util.CreateIntValidator(&logFileMaxAge, &m.config.LogFile.MaxAge)),
			huh.NewConfirm().
				Title("Log File Local Time").
				Description("Used for timestamping files. Default is UTC time.").
				Value(&m.config.LogFile.LocalTime),
			huh.NewConfirm().
				Title("Log File Compress").
				Value(&m.config.LogFile.Compress),
		),
	).WithTheme(views.GetCustomTheme()).WithHeight(20)
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) ValidateField() {
	// Validate the current field before submitting the form.
	m.form.NextField()
	m.form.PrevField()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			m.ValidateField()
			if len(m.form.Errors()) > 0 {
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c":
			m.config = nil
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
		m.quitting = true
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	helpView := ""
	if len(m.form.Errors()) == 0 {
		helpView = views.DefaultRowDataStyle.Render(" • " + m.help.ShortHelpView([]key.Binding{m.keymap.submit}))
	}

	// TODO: once huh is updated to properly focus fields, add alt screen titles
	// return views.GetAltScreenTitle("SERVER CONFIGURATION") + m.form.View() + helpView
	return m.form.View() + helpView
}

func ConfigurationForm(config *runner.Config) (*runner.Config, error) {
	m := NewModel(config)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()

	if err != nil {
		return nil, err
	}

	if m, ok := p.(Model); ok && m.config != nil {
		return m.config, nil
	}

	return nil, errors.New("no changes were made")
}
