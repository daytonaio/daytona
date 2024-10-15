// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	info_view "github.com/daytonaio/daytona/pkg/views/workspace/info"
)

type RowData struct {
	Name       string
	Repository string
	Target     string
	Status     string
	Created    string
	Branch     string
}

func ListWorkspaces(workspaceList []apiclient.WorkspaceDTO, specifyGitProviders bool, verbose bool, activeProfileName string) {
	SortWorkspaces(&workspaceList, verbose)

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
			row = getRowFromRowData(RowData{Name: workspace.Name}, true)
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

	footer := lipgloss.NewStyle().Foreground(views.LightGray).Render(views.GetListFooter(activeProfileName, &views.Padding{}))

	table := views_util.GetTableView(data, headers, &footer, func() {
		renderUnstyledList(workspaceList)
	})

	fmt.Println(table)
}

func renderUnstyledList(workspaceList []apiclient.WorkspaceDTO) {
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
		views.DefaultRowDataStyle.Render(views.GetBranchNameLabel(rowData.Branch)),
	}

	if rowData.Status != "" {
		row[3] = fmt.Sprintf("%s %s", state, views.DefaultRowDataStyle.Render(fmt.Sprintf("(%s)", rowData.Status)))
	}

	return row
}

func SortWorkspaces(workspaceList *[]apiclient.WorkspaceDTO, verbose bool) {
	if verbose {
		sort.Slice(*workspaceList, func(i, j int) bool {
			ws1 := (*workspaceList)[i]
			ws2 := (*workspaceList)[j]
			if ws1.Info == nil || ws2.Info == nil || ws1.Info.Projects == nil || ws2.Info.Projects == nil {
				return true
			}
			if len(ws1.Info.Projects) == 0 || len(ws2.Info.Projects) == 0 {
				return true
			}
			return ws1.Info.Projects[0].Created > ws2.Info.Projects[0].Created
		})
		return
	}

	sort.Slice(*workspaceList, func(i, j int) bool {
		ws1 := (*workspaceList)[i]
		ws2 := (*workspaceList)[j]
		if len(ws1.Projects) == 0 || len(ws2.Projects) == 0 || ws1.Projects[0].State == nil || ws2.Projects[0].State == nil {
			return true
		}
		return ws1.Projects[0].State.Uptime < ws2.Projects[0].State.Uptime
	})

}

func getWorkspaceTableRowData(workspace apiclient.WorkspaceDTO, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", "", ""}
	rowData.Name = workspace.Name + views_util.AdditionalPropertyPadding
	if len(workspace.Projects) > 0 {
		rowData.Repository = util.GetRepositorySlugFromUrl(workspace.Projects[0].Repository.Url, specifyGitProviders)
		rowData.Branch = workspace.Projects[0].Repository.Branch
	}

	rowData.Target = workspace.Target + views_util.AdditionalPropertyPadding

	if workspace.Info != nil && workspace.Info.Projects != nil && len(workspace.Info.Projects) > 0 {
		rowData.Created = util.FormatTimestamp(workspace.Info.Projects[0].Created)
	}
	if len(workspace.Projects) > 0 && workspace.Projects[0].State != nil && workspace.Projects[0].State.Uptime > 0 {
		rowData.Status = util.FormatUptime(workspace.Projects[0].State.Uptime)
	}
	return &rowData
}

func getProjectTableRowData(workspaceDTO apiclient.WorkspaceDTO, project apiclient.Project, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", "", ""}
	rowData.Name = " └ " + project.Name

	rowData.Repository = util.GetRepositorySlugFromUrl(project.Repository.Url, specifyGitProviders)
	rowData.Branch = project.Repository.Branch

	rowData.Target = project.Target + views_util.AdditionalPropertyPadding

	if project.State != nil && project.State.Uptime > 0 {
		rowData.Status = util.FormatUptime(project.State.Uptime)
	}

	if workspaceDTO.Info == nil || workspaceDTO.Info.Projects == nil {
		return &rowData
	}

	for _, projectInfo := range workspaceDTO.Info.Projects {
		if projectInfo.Name == project.Name {
			rowData.Created = util.FormatTimestamp(projectInfo.Created)
			break
		}
	}

	return &rowData
}
