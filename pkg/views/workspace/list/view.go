// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/types"
	"golang.org/x/term"
)

var defaultColumnWidth = 12
var columnPadding = 3
var timeLayout = "2006-01-02T15:04:05.999999999Z"

type RowData struct {
	WorkspaceName string
	Repository    string
	Target        string
	Created       string
	Status        string
}

type model struct {
	table       table.Model
	selectable  bool
	initialRows []table.Row
}

var columns = []table.Column{
	{Title: "WORKSPACE", Width: defaultColumnWidth},
	{Title: "REPOSITORY", Width: defaultColumnWidth},
	{Title: "TARGET", Width: defaultColumnWidth},
	{Title: "CREATED", Width: defaultColumnWidth},
	{Title: "STATUS", Width: defaultColumnWidth},
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

func renderWorkspaceList(workspaceList []serverapiclient.Workspace, specifyGitProviders bool) model {
	rows := []table.Row{}
	var row table.Row
	var rowData RowData

	sortWorkspaces(&workspaceList)

	for _, workspace := range workspaceList {
		if len(workspace.Projects) == 1 {
			rowData = getWorkspaceTableRowData(workspace, specifyGitProviders)
			adjustColumsFormatting(rowData)
			row = table.Row{rowData.WorkspaceName, rowData.Repository, rowData.Target, rowData.Created, rowData.Status}
			rows = append(rows, row)
		} else {
			row = table.Row{*workspace.Name, "", "", "", "", ""}
			rows = append(rows, row)
			for _, project := range workspace.Projects {
				rowData = getProjectTableRowData(workspace, project, specifyGitProviders)
				adjustColumsFormatting(rowData)
				row = table.Row{rowData.WorkspaceName, rowData.Repository, rowData.Target, rowData.Created, rowData.Status}
				rows = append(rows, row)
			}
		}
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	adjustedRows, adjustedCols := getRowsAndCols(width, rows)

	return model{
		table:       getTable(adjustedRows, adjustedCols, 0),
		selectable:  false,
		initialRows: rows,
	}
}

func sortWorkspaces(workspaceList *[]serverapiclient.Workspace) {
	sort.Slice(*workspaceList, func(i, j int) bool {
		ws1 := (*workspaceList)[i]
		ws2 := (*workspaceList)[j]
		if ws1.Info == nil || ws2.Info == nil || ws1.Info.Projects == nil || ws2.Info.Projects == nil {
			return false
		}
		if len(ws1.Info.Projects) == 0 {
			return false
		}
		if len(ws2.Info.Projects) == 0 {
			return true
		}
		return *ws1.Info.Projects[0].Created > *ws2.Info.Projects[0].Created
	})
}

func adjustColumsFormatting(rowData RowData) {
	adjustColumnWidth("WORKSPACE", rowData)
	adjustColumnWidth("REPOSITORY", rowData)
	adjustColumnWidth("TARGET", rowData)
	adjustColumnWidth("CREATED", rowData)
	adjustColumnWidth("STATUS", rowData)
}

func adjustColumnWidth(title string, rowData RowData) {
	var column *table.Column
	for i, col := range columns {
		if col.Title == title {
			column = &columns[i]
			break
		}
	}
	currentField := ""
	switch title {
	case "WORKSPACE":
		currentField = rowData.WorkspaceName
	case "REPOSITORY":
		currentField = rowData.Repository
	case "TARGET":
		currentField = rowData.Target
	case "CREATED":
		currentField = rowData.Created
	case "STATUS":
		currentField = rowData.Status
	}

	if len(currentField) > column.Width {
		column.Width = len(currentField) + columnPadding
	}
}

func getWorkspaceTableRowData(workspace serverapiclient.Workspace, specifyGitProviders bool) RowData {
	rowData := RowData{}
	if workspace.Name != nil {
		rowData.WorkspaceName = *workspace.Name
	}
	if workspace.Projects != nil && len(workspace.Projects) > 0 && workspace.Projects[0].Repository != nil {
		rowData.Repository = getRepositorySlugFromUrl(*workspace.Projects[0].Repository.Url, specifyGitProviders)
	}
	if workspace.Target != nil {
		rowData.Target = *workspace.Target
	}
	if workspace.Info != nil && workspace.Info.Projects != nil && len(workspace.Info.Projects) > 0 && workspace.Info.Projects[0].Created != nil {
		rowData.Created = formatCreatedTime(*workspace.Info.Projects[0].Created)
	}
	if workspace.Info != nil && workspace.Info.Projects != nil && len(workspace.Info.Projects) > 0 && workspace.Info.Projects[0].Started != nil {
		rowData.Status = formatStatusTime(*workspace.Info.Projects[0].Started)
	}
	return rowData
}

func getProjectTableRowData(workspace serverapiclient.Workspace, project serverapiclient.Project, specifyGitProviders bool) RowData {
	var currentProjectInfo *types.ProjectInfo

	for _, projectInfo := range workspace.Info.Projects {
		if *projectInfo.Name == *project.Name {
			currentProjectInfo = &types.ProjectInfo{
				Name:    *projectInfo.Name,
				Created: *projectInfo.Created,
				Started: *projectInfo.Started,
			}
			break
		}
	}

	if currentProjectInfo == nil {
		currentProjectInfo = &types.ProjectInfo{
			Name:    *project.Name,
			Created: "/",
			Started: "/",
		}
	}

	rowData := RowData{}
	if project.Name != nil {
		rowData.WorkspaceName = " └ " + *project.Name
	}
	if project.Repository != nil && project.Repository.Url != nil {
		rowData.Repository = getRepositorySlugFromUrl(*project.Repository.Url, specifyGitProviders)
	}
	if project.Target != nil {
		rowData.Target = *project.Target
	}
	rowData.Created = formatCreatedTime(currentProjectInfo.Created)
	rowData.Status = formatStatusTime(currentProjectInfo.Started)
	return rowData
}

func getRepositorySlugFromUrl(url string, specifyGitProviders bool) string {
	if url == "" {
		return "/"
	}
	url = strings.TrimSuffix(url, "/")

	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return ""
	}

	if specifyGitProviders {
		return parts[len(parts)-3] + "/" + parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}

	return parts[len(parts)-2] + "/" + parts[len(parts)-1]
}

func formatCreatedTime(input string) string {
	t, err := time.Parse(timeLayout, input)
	if err != nil {
		return "/"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "< 1 minute ago"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func formatStatusTime(input string) string {
	t, err := time.Parse(timeLayout, input)
	if err != nil {
		return "stopped"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "up < 1 minute"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "up 1 minute"
		}
		return fmt.Sprintf("up %d minutes", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "up 1 hour"
		}
		return fmt.Sprintf("up %d hours", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "up 1 day"
		}
		return fmt.Sprintf("up %d days", days)
	}
}

func ListWorkspaces(workspaceList []serverapiclient.Workspace, specifyGitProviders bool) {
	modelInstance := renderWorkspaceList(workspaceList, specifyGitProviders)

	_, err := tea.NewProgram(modelInstance).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	fmt.Println()
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
