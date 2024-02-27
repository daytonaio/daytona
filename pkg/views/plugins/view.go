// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"fmt"
	"os"
	"strings"

	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type PluginType string

type PluginViewDTO struct {
	Name    string
	Version string
	Type    PluginType
}

const (
	PluginTypeProvisioner  PluginType = "Provisioner"
	PluginTypeAgentService PluginType = "Agent Service"
)

var columns = []table.Column{
	{Title: "Name", Width: 20},
	{Title: "Version", Width: 20},
	{Title: "Type", Width: 20},
}

type model struct {
	table          table.Model
	selectedPlugin *PluginViewDTO
	selectable     bool
	initialRows    []table.Row
}

func (m model) Init() tea.Cmd {
	if !m.selectable {
		return tea.Quit
	}

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rows, cols := getRowsAndCols(msg.Width, m.initialRows)
		m.table = getTable(rows, cols, m.selectable, m.table.Cursor())
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			m.selectedPlugin = nil
			return m, tea.Quit
		case "enter":
			m.selectedPlugin = &PluginViewDTO{
				Name:    m.table.SelectedRow()[0],
				Version: m.table.SelectedRow()[1],
				Type:    PluginType(m.table.SelectedRow()[2]),
			}
			return m, tea.Quit
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	Padding(0, 1).MarginBottom(1)

func (m model) View() string {
	return baseStyle.Render(m.table.View())
}

func renderPluginsList(plugins []PluginViewDTO, selectable bool) model {
	rows := []table.Row{}
	selectedPlugin := &plugins[0]

	for _, plugin := range plugins {
		row := table.Row{plugin.Name, plugin.Version, string(plugin.Type)}

		rows = append(rows, row)
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))

	adjustedRows, adjustedCols := getRowsAndCols(width, rows)

	return model{
		table:          getTable(adjustedRows, adjustedCols, selectable, 0),
		selectedPlugin: selectedPlugin,
		selectable:     selectable,
		initialRows:    rows,
	}
}

func GetPluginFromPrompt(plugins []PluginViewDTO, title string) *PluginViewDTO {
	util.RenderMainTitle(title)

	if len(plugins) == 0 {
		fmt.Println("No plugins found")
		return nil
	}

	modelInstance := renderPluginsList(plugins, true)

	m, err := tea.NewProgram(modelInstance).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	selectedPlugin := m.(model).selectedPlugin

	lipgloss.DefaultRenderer().Output().ClearLines(strings.Count(modelInstance.View(), "\n") + 2)

	return selectedPlugin
}

func ListPlugins(plugins []PluginViewDTO) {
	util.RenderMainTitle("Plugins")

	if len(plugins) == 0 {
		fmt.Println("No plugins found")
		return
	}

	modelInstance := renderPluginsList(plugins, false)

	_, err := tea.NewProgram(modelInstance).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func getTable(rows []table.Row, cols []table.Column, selectable bool, activeRow int) table.Model {
	var t table.Model

	if selectable {
		t = table.New(
			table.WithColumns(cols),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(len(rows)),
		)
	} else {
		t = table.New(
			table.WithColumns(cols),
			table.WithRows(rows),
			table.WithHeight(len(rows)),
		)
	}

	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		AlignHorizontal(lipgloss.Left)

	if selectable {
		style.Selected = style.Selected.
			Foreground(lipgloss.Color(views.White.Dark)).
			Background(lipgloss.Color(views.Green.Dark)).
			Bold(false)
	} else {
		style.Selected = style.Selected.
			Foreground(style.Cell.GetForeground()).
			Background(style.Cell.GetBackground()).
			Bold(false)
	}

	t.SetStyles(style)
	t.SetCursor(activeRow)

	return t
}

func getRowsAndCols(width int, initialRows []table.Row) ([]table.Row, []table.Column) {
	colWidth := 0
	cols := []table.Column{}

	for _, col := range columns {
		if colWidth+col.Width > width {
			break
		}

		colWidth += col.Width
		cols = append(cols, col)
	}

	rows := []table.Row{}
	for _, row := range initialRows {
		rows = append(rows, row[:len(cols)])
	}

	return rows, cols
}
