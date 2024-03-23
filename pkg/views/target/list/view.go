// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/util"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var columns = []table.Column{
	{Title: "Provider", Width: 24},
	{Title: "Target", Width: 24},
	{Title: "Options", Width: 70},
}

type model struct {
	table       table.Model
	initialRows []table.Row
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
	BorderStyle(lipgloss.RoundedBorder()).
	Padding(0, 1).MarginBottom(1)

func (m model) View() string {
	return baseStyle.Render(m.table.View())
}

func getTable(rows []table.Row, cols []table.Column, activeRow int) table.Model {
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)),
	)

	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.NormalBorder()).
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

func ListTargets(targets []serverapiclient.ProviderTarget) {
	util.RenderMainTitle("TARGETS")

	m := renderTargetList(targets)

	_, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func renderTargetList(targets []serverapiclient.ProviderTarget) model {
	rows := []table.Row{}

	sortTargets(&targets)

	for _, target := range targets {
		optionsString := *target.Options
		optionsRows := []table.Row{}

		parts := strings.Split(optionsString, "\n")
		for i, part := range parts {
			if i == 0 {
				optionsRows = append(optionsRows, table.Row{
					*target.ProviderInfo.Name,
					*target.Name,
					part,
				})
			} else {
				optionsRows = append(optionsRows, table.Row{"", "", part})
			}
		}

		rows = append(rows, optionsRows...)
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	adjustedRows, adjustedCols := getRowsAndCols(width, rows)

	return model{
		table:       getTable(adjustedRows, adjustedCols, 0),
		initialRows: rows,
	}

}

func sortTargets(targets *[]serverapiclient.ProviderTarget) {
	sort.Slice(*targets, func(i, j int) bool {
		t1 := (*targets)[i]
		t2 := (*targets)[j]
		return *t1.ProviderInfo.Name < *t2.ProviderInfo.Name
	})
}
