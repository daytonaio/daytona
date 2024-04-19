// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"golang.org/x/term"
)

var defaultColumnWidth = 12
var columnPadding = 3

type rowData struct {
	Server   string
	Username string
	Password string
}

type model struct {
	table       table.Model
	initialRows []table.Row
}

var columns = []table.Column{
	{Title: "SERVER", Width: defaultColumnWidth},
	{Title: "USERNAME", Width: defaultColumnWidth},
	{Title: "PASSWORD", Width: defaultColumnWidth},
}

func (m model) Init() tea.Cmd {
	return tea.Quit
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rows, cols := getRowsAndCols(msg.Width, m.initialRows)
		m.table = getTable(rows, cols, m.table.Cursor())
		return m, nil
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.HiddenBorder())

func (m model) View() string {
	return baseStyle.Render(m.table.View())
}

func renderRegistryList(registryList []serverapiclient.ContainerRegistry) model {
	rows := []table.Row{}
	var row table.Row
	var rowData rowData

	for _, registry := range registryList {
		rowData.Server = *registry.Server
		rowData.Username = *registry.Username
		rowData.Password = *registry.Password

		adjustColumsFormatting(rowData)
		row = table.Row{rowData.Server, rowData.Username, rowData.Password}
		rows = append(rows, row)
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	adjustedRows, adjustedCols := getRowsAndCols(width, rows)

	return model{
		table:       getTable(adjustedRows, adjustedCols, 0),
		initialRows: rows,
	}
}

func adjustColumsFormatting(rowData rowData) {
	adjustColumnWidth("SERVER", rowData)
	adjustColumnWidth("USERNAME", rowData)
	adjustColumnWidth("PASSWORD", rowData)
}

func adjustColumnWidth(title string, rowData rowData) {
	var column *table.Column
	for i, col := range columns {
		if col.Title == title {
			column = &columns[i]
			break
		}
	}
	currentField := ""
	switch title {
	case "SERVER":
		currentField = rowData.Server
	case "USERNAME":
		currentField = rowData.Username
	case "PASSWORD":
		currentField = rowData.Password
	}

	if len(currentField) > column.Width {
		column.Width = len(currentField) + columnPadding
	}
}

func ListRegistries(registryList []serverapiclient.ContainerRegistry) error {
	modelInstance := renderRegistryList(registryList)

	_, err := tea.NewProgram(modelInstance).Run()
	if err != nil {
		return err
	}
	fmt.Println()
	return nil
}

func getTable(rows []table.Row, cols []table.Column, activeRow int) table.Model {
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(len(rows)),
	)

	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.HiddenBorder()).
		BorderBottom(true).
		AlignHorizontal(lipgloss.Left)

	style.Selected = style.Selected.
		Foreground(style.Cell.GetForeground()).
		Background(style.Cell.GetBackground()).
		Bold(false)

	t.SetStyles(style)
	t.SetCursor(activeRow)

	return t
}

func getRowsAndCols(width int, initialRows []table.Row) ([]table.Row, []table.Column) {
	colWidth := 0
	cols := []table.Column{}

	for i, col := range columns {
		// keep columns length in sync with initialRows
		if i >= len(initialRows[0]) {
			break
		}

		if colWidth+col.Width > width {
			break
		}

		colWidth += col.Width
		cols = append(cols, col)
	}

	rows := make([]table.Row, len(initialRows))

	for i, row := range initialRows {
		if len(row) >= len(cols) {
			rows[i] = row[:len(cols)]
		} else {
			rows[i] = row
		}
	}
	return rows, cols
}
