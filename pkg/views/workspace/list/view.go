// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	info_view "github.com/daytonaio/daytona/pkg/views/workspace/info"
	"golang.org/x/term"
)

type RowData struct {
	Name       string
	Repository string
	Target     string
	Status     string
	Created    string
	Branch     string
}

func ListWorkspaces(workspaceList []serverapiclient.WorkspaceDTO, specifyGitProviders bool, verbose bool) {
	sortWorkspaces(&workspaceList, verbose)

	re := lipgloss.NewRenderer(os.Stdout)

	headers := []string{"Workspace", "Repository", "Target", "Status", "Created", "Branch"}

	data := [][]string{}

	for _, workspace := range workspaceList {
		var rowData *RowData
		var row []string

		if len(workspace.Projects) == 1 {
			rowData = getWorkspaceTableRowData(workspace, specifyGitProviders)
			row = getRowFromRowData(*rowData, false)
			data = append(data, row)
		} else {
			row = getRowFromRowData(RowData{Name: *workspace.Name}, true)
			data = append(data, row)
			for _, project := range workspace.Projects {
				rowData = getProjectTableRowData(workspace, project, specifyGitProviders)
				if rowData == nil {
					continue
				}
				row = getRowFromRowData(*rowData, false)
				data = append(data, row)
			}
		}
	}

	if !verbose {
		headers = headers[:len(headers)-2]
		for value := range data {
			data[value] = data[value][:len(data[value])-2]
		}
	} else {
		// Temporarily hiding the branch column
		headers = headers[:len(headers)-1]
		for value := range data {
			data[value] = data[value][:len(data[value])-1]
		}
	}

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(data)
		return
	}

	breakpointWidth := views.GetContainerBreakpointWidth(terminalWidth)

	minWidth := views_util.GetTableMinimumWidth(data)

	if breakpointWidth == 0 || minWidth > breakpointWidth {
		renderUnstyledList(workspaceList)
		return
	}

	t := table.New().
		Headers(headers...).
		Rows(data...).
		BorderStyle(re.NewStyle().Foreground(views.LightGray)).
		BorderRow(false).BorderColumn(false).BorderLeft(false).BorderRight(false).BorderTop(false).BorderBottom(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return views.TableHeaderStyle
			}
			return views.BaseCellStyle
		}).Width(breakpointWidth - 2*views.BaseTableStyleHorizontalPadding - 1)

	fmt.Println(views.BaseTableStyle.Render(t.String()))
}

func renderUnstyledList(workspaceList []serverapiclient.WorkspaceDTO) {
	for _, workspace := range workspaceList {
		info_view.Render(&workspace, "", true)

		if workspace.Id != workspaceList[len(workspaceList)-1].Id {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}

	}
}

func getRowFromRowData(rowData RowData, isMultiProjectAccordion bool) []string {
	var state string
	if rowData.Status == "" {
		state = views.InactiveStyle.Render("STOPPED")
	} else {
		state = views.ActiveStyle.Render("RUNNING")
	}

	if isMultiProjectAccordion {
		return []string{rowData.Name, "", "", "", "", ""}
	}

	row := []string{
		views.NameStyle.Render(rowData.Name),
		views.DefaultRowDataStyle.Render(rowData.Repository),
		views.DefaultRowDataStyle.Render(rowData.Target),
		state,
		views.DefaultRowDataStyle.Render(rowData.Created),
		views.DefaultRowDataStyle.Render(rowData.Branch),
	}

	if rowData.Status != "" {
		row[3] = fmt.Sprintf("%s %s", state, views.DefaultRowDataStyle.Render(fmt.Sprintf("(%s)", rowData.Status)))
	}

	return row
}

func sortWorkspaces(workspaceList *[]serverapiclient.WorkspaceDTO, verbose bool) {
	if verbose {
		sort.Slice(*workspaceList, func(i, j int) bool {
			ws1 := (*workspaceList)[i]
			ws2 := (*workspaceList)[j]
			if ws1.Info == nil || ws2.Info == nil || ws1.Info.Projects == nil || ws2.Info.Projects == nil {
				return false
			}
			if len(ws1.Info.Projects) == 0 || len(ws2.Info.Projects) == 0 || ws1.Info.Projects[0].Created == nil || ws2.Info.Projects[0].Created == nil {
				return false
			}
			return *ws1.Info.Projects[0].Created > *ws2.Info.Projects[0].Created
		})
		return
	}

	sort.Slice(*workspaceList, func(i, j int) bool {
		ws1 := (*workspaceList)[i]
		ws2 := (*workspaceList)[j]
		if len(ws1.Projects) == 0 || len(ws2.Projects) == 0 || ws1.Projects[0].State == nil || ws2.Projects[0].State == nil || ws1.Projects[0].State.Uptime == nil || ws2.Projects[0].State.Uptime == nil {
			return true
		}
		return *ws1.Projects[0].State.Uptime < *ws2.Projects[0].State.Uptime
	})

}

func getWorkspaceTableRowData(workspace serverapiclient.WorkspaceDTO, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", "", ""}
	if workspace.Name != nil {
		rowData.Name = *workspace.Name + views_util.AdditionalPropertyPadding
	}
	if workspace.Projects != nil && len(workspace.Projects) > 0 && workspace.Projects[0].Repository != nil {
		rowData.Repository = getRepositorySlugFromUrl(*workspace.Projects[0].Repository.Url, specifyGitProviders)
		if workspace.Projects[0].Repository.Branch != nil {
			rowData.Branch = *workspace.Projects[0].Repository.Branch
		}
	}
	if workspace.Target != nil {
		rowData.Target = *workspace.Target + views_util.AdditionalPropertyPadding
	}
	if workspace.Info != nil && workspace.Info.Projects != nil && len(workspace.Info.Projects) > 0 && workspace.Info.Projects[0].Created != nil {
		rowData.Created = util.FormatCreatedTime(*workspace.Info.Projects[0].Created)
	}
	if len(workspace.Projects) > 0 && workspace.Projects[0].State != nil && workspace.Projects[0].State.Uptime != nil {
		rowData.Status = util.FormatUptime(*workspace.Projects[0].State.Uptime)
	}
	return &rowData
}

func getProjectTableRowData(workspaceDTO serverapiclient.WorkspaceDTO, project serverapiclient.Project, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", "", ""}
	if project.Name != nil {
		rowData.Name = " └ " + *project.Name
	}
	if project.Repository != nil && project.Repository.Url != nil {
		rowData.Repository = getRepositorySlugFromUrl(*project.Repository.Url, specifyGitProviders)
		if project.Repository.Branch != nil {
			rowData.Branch = *project.Repository.Branch
		}
	}
	if project.Target != nil {
		rowData.Target = *project.Target + views_util.AdditionalPropertyPadding
	}
	if project.State != nil && project.State.Uptime != nil {
		rowData.Status = util.FormatUptime(*project.State.Uptime)
	}

	if workspaceDTO.Info == nil || workspaceDTO.Info.Projects == nil {
		return &rowData
	}

	for _, projectInfo := range workspaceDTO.Info.Projects {
		if *projectInfo.Name == *project.Name {
			rowData.Created = util.FormatCreatedTime(*projectInfo.Created)
			break
		}
	}

	return &rowData
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
