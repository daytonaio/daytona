// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"
	"os"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var NewProfileId = "+"

var columns = []table.Column{
	{Title: "ID", Width: 10},
	{Title: "Name", Width: 20},
	{Title: "Active", Width: 10},
	{Title: "API URL", Width: 30},
}

type model struct {
	table             table.Model
	selectedProfileId string
	selectable        bool
	initialRows       []table.Row
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
			m.selectedProfileId = ""
			return m, tea.Quit
		case "enter":
			m.selectedProfileId = m.table.SelectedRow()[0]
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

func renderProfileList(profileList []config.Profile, activeProfileId string, selectable bool) model {
	rows := []table.Row{}
	activeProfileRow := 0
	for i, profile := range profileList {
		row := table.Row{profile.Id, profile.Name, fmt.Sprintf("%t", profile.Id == activeProfileId), profile.Api.Url}

		switch profile.Id {
		case NewProfileId:
			row = table.Row{profile.Id, profile.Name, "", ""}
		}

		if profile.Id == activeProfileId {
			activeProfileRow = i
		}

		rows = append(rows, row)
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))

	adjustedRows, adjustedCols := getRowsAndCols(width, rows)

	return model{
		table:             getTable(adjustedRows, adjustedCols, selectable, activeProfileRow),
		selectedProfileId: activeProfileId,
		selectable:        selectable,
		initialRows:       rows,
	}
}

func GetProfileIdFromPrompt(profileList []config.Profile, activeProfileId, title string, withCreateOption bool) string {
	util.RenderMainTitle(title)

	withNewProfile := profileList

	if withCreateOption {
		withNewProfile = append(withNewProfile, config.Profile{
			Id:   NewProfileId,
			Name: "Add new profile",
		})
	}

	modelInstance := renderProfileList(withNewProfile, activeProfileId, true)

	m, err := tea.NewProgram(modelInstance).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	profileId := m.(model).selectedProfileId

	lipgloss.DefaultRenderer().Output().ClearLines(strings.Count(modelInstance.View(), "\n") + 2)

	return profileId
}

func ListProfiles(profileList []config.Profile, activeProfileId string) {
	util.RenderMainTitle("PROFILES")

	modelInstance := renderProfileList(profileList, activeProfileId, false)

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
			Background(lipgloss.Color(views.DimmedGreen.Dark)).
			Bold(true)
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
