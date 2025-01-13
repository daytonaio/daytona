// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"
)

type Model struct {
	form     *huh.Form
	quitting bool
	config   *apiclient.ServerConfig
	keymap   keymap
	help     help.Model
}

type keymap struct {
	submit key.Binding
}

func NewModel(config *apiclient.ServerConfig) Model {
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
	apiPortView := strconv.Itoa(int(m.config.GetApiPort()))
	headscalePortView := strconv.Itoa(int(m.config.GetHeadscalePort()))
	frpsPortView := strconv.Itoa(int(m.config.Frps.GetPort()))
	localBuilderRegistryPort := strconv.Itoa(int(m.config.GetLocalBuilderRegistryPort()))

	logFileMaxSize := strconv.Itoa(int(m.config.LogFile.MaxSize))
	logFileMaxBackups := strconv.Itoa(int(m.config.LogFile.MaxBackups))
	logFileMaxAge := strconv.Itoa(int(m.config.LogFile.MaxAge))

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Registry URL").
				Value(&m.config.RegistryUrl),
			huh.NewInput().
				Title("Server Download URL").
				Value(&m.config.ServerDownloadUrl),
			huh.NewInput().
				Title("Samples Index URL").
				Description("Leave empty to disable samples").
				Value(m.config.SamplesIndexUrl),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Default Workspace Image").
				Value(&m.config.DefaultWorkspaceImage),
			huh.NewInput().
				Title("Default Workspace User").
				Value(&m.config.DefaultWorkspaceUser),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Builder Image").
				Description("Image dependencies: docker, socat, git, @devcontainers/cli (node package)").
				Value(&m.config.BuilderImage),
			huh.NewInput().
				Title("Builder Registry Server").
				Description("Add container registry credentials to the server by adding them as environment variables using `daytona env set`").
				Value(&m.config.BuilderRegistryServer),
			huh.NewInput().
				Title("Build Image Namespace").
				Description("Namespace to be used when tagging and pushing build images").
				Value(m.config.BuildImageNamespace),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Local Builder Registry Port").
				Value(&localBuilderRegistryPort).
				Validate(util.CreateServerPortValidator(m.config, &localBuilderRegistryPort, &m.config.LocalBuilderRegistryPort)),
			huh.NewInput().
				Title("Local Builder Registry Image").
				Value(&m.config.LocalBuilderRegistryImage),
			huh.NewConfirm().
				Title("Local Runner Disabled").
				Description("Disables the local runner").
				Value(m.config.LocalRunnerDisabled),
		).WithHideFunc(func() bool {
			return m.config.BuilderRegistryServer != "local"
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("API Port").
				Value(&apiPortView).
				Validate(util.CreateServerPortValidator(m.config, &apiPortView, &m.config.ApiPort)),
			huh.NewInput().
				Title("Headscale Port").
				Value(&headscalePortView).
				Validate(util.CreateServerPortValidator(m.config, &headscalePortView, &m.config.HeadscalePort)),
			huh.NewInput().
				Title("Binaries Path").
				Description("Directory will be created if it does not exist").
				Value(&m.config.BinariesPath),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Log File Path").
				Description("File will be created if it does not exist").
				Value(&m.config.LogFile.Path).
				Validate(func(s string) error {
					_, err := os.Stat(s)
					if os.IsNotExist(err) {
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
				Value(m.config.LogFile.LocalTime),
			huh.NewConfirm().
				Title("Log File Compress").
				Value(m.config.LogFile.Compress),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Frps Domain").
				Value(&m.config.Frps.Domain),
			huh.NewInput().
				Title("Frps Port").
				Value(&frpsPortView).
				Validate(util.CreateServerPortValidator(m.config, &frpsPortView, &m.config.Frps.Port)),
			huh.NewInput().
				Title("Frps Protocol").
				Value(&m.config.Frps.Protocol),
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

func ConfigurationForm(config *apiclient.ServerConfig) (*apiclient.ServerConfig, error) {
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
