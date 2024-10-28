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

func NewModel(config *apiclient.ServerConfig, containerRegistries []apiclient.ContainerRegistry) Model {
	m := Model{
		config: config,
		help:   help.New(),
		keymap: keymap{
			submit: key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "submit")),
		},
	}

	m.form = m.createForm(containerRegistries)
	return m
}
func (m *Model) createForm(containerRegistries []apiclient.ContainerRegistry) *huh.Form {
	apiPortView := strconv.Itoa(int(m.config.GetApiPort()))
	headscalePortView := strconv.Itoa(int(m.config.GetHeadscalePort()))
	frpsPortView := strconv.Itoa(int(m.config.Frps.GetPort()))
	localBuilderRegistryPort := strconv.Itoa(int(m.config.GetLocalBuilderRegistryPort()))

	builderContainerRegistryOptions := []huh.Option[string]{{
		Key:   "Local registry managed by Daytona",
		Value: "local",
	}}
	for _, cr := range containerRegistries {
		builderContainerRegistryOptions = append(builderContainerRegistryOptions, huh.Option[string]{Key: cr.Server, Value: cr.Server})
	}

	logFileMaxSize := strconv.Itoa(int(m.config.LogFile.MaxSize))
	logFileMaxBackups := strconv.Itoa(int(m.config.LogFile.MaxBackups))
	logFileMaxAge := strconv.Itoa(int(m.config.LogFile.MaxAge))

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Providers Directory").
				Description("Directory will be created if it does not exist").
				Value(&m.config.ProvidersDir),
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
			huh.NewSelect[string]().
				Title("Builder Registry").
				Description("To add options, add a container registry with 'daytona cr set'").
				Options(
					builderContainerRegistryOptions...,
				).
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
				Validate(createPortValidator(m.config, &localBuilderRegistryPort, &m.config.LocalBuilderRegistryPort)),
			huh.NewInput().
				Title("Local Builder Registry Image").
				Value(&m.config.LocalBuilderRegistryImage),
		).WithHideFunc(func() bool {
			return m.config.BuilderRegistryServer != "local"
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("API Port").
				Value(&apiPortView).
				Validate(createPortValidator(m.config, &apiPortView, &m.config.ApiPort)),
			huh.NewInput().
				Title("Headscale Port").
				Value(&headscalePortView).
				Validate(createPortValidator(m.config, &headscalePortView, &m.config.HeadscalePort)),
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
				Validate(createIntValidator(&logFileMaxSize, &m.config.LogFile.MaxSize)),
			huh.NewInput().
				Title("Log File Max Backups").
				Value(&logFileMaxBackups).
				Validate(createIntValidator(&logFileMaxBackups, &m.config.LogFile.MaxBackups)),
			huh.NewInput().
				Title("Log File Max Age").
				Description("In days").
				Value(&logFileMaxAge).
				Validate(createIntValidator(&logFileMaxAge, &m.config.LogFile.MaxAge)),
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
				Validate(createPortValidator(m.config, &frpsPortView, &m.config.Frps.Port)),
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

func ConfigurationForm(config *apiclient.ServerConfig, containerRegistries []apiclient.ContainerRegistry) (*apiclient.ServerConfig, error) {
	m := NewModel(config, containerRegistries)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()

	if err != nil {
		return nil, err
	}

	if m, ok := p.(Model); ok && m.config != nil {
		return m.config, nil
	}

	return nil, errors.New("no changes were made")
}

func createPortValidator(config *apiclient.ServerConfig, portView *string, port *int32) func(string) error {
	return func(string) error {
		validatePort, err := strconv.Atoi(*portView)
		if err != nil {
			return errors.New("failed to parse port")
		}
		if validatePort < 0 || validatePort > 65535 {
			return errors.New("port out of range")
		}
		*port = int32(validatePort)

		if config.ApiPort == config.HeadscalePort {
			return errors.New("port conflict")
		}

		return nil
	}
}

func createIntValidator(viewValue *string, value *int32) func(string) error {
	return func(string) error {
		validateInt, err := strconv.Atoi(*viewValue)
		if err != nil {
			return errors.New("failed to parse int")
		}

		if validateInt <= 0 {
			return errors.New("int out of range")
		}

		*value = int32(validateInt)

		return nil
	}
}
