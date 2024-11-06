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
	TargetName string
	Status     string
	Created    string
	Branch     string
}

func ListWorkspaces(workspaceList []apiclient.WorkspaceDTO, specifyGitProviders bool, verbose bool, activeProfileName string) {
	if len(workspaceList) == 0 {
		views_util.NotifyEmptyWorkspaceList(true)
		return
	}

	SortWorkspaces(&workspaceList, verbose)

	headers := []string{"Workspace", "Repository", "Target", "Status", "Created", "Branch"}

	data := [][]string{}

	for _, target := range workspaceList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(target, specifyGitProviders)
		row = getRowFromRowData(*rowData, false)
		data = append(data, row)
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
	for _, target := range workspaceList {
		info_view.Render(&target, "", true)

		if target.Id != workspaceList[len(workspaceList)-1].Id {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}

	}
}

func getRowFromRowData(rowData RowData, isMultiWorkspaceAccordion bool) []string {
	var state string
	if rowData.Status == "" {
		state = views.InactiveStyle.Render("STOPPED")
	} else {
		state = views.ActiveStyle.Render("RUNNING")
	}

	if isMultiWorkspaceAccordion {
		return []string{rowData.Name, "", "", "", "", ""}
	}

	row := []string{
		views.NameStyle.Render(rowData.Name),
		views.DefaultRowDataStyle.Render(rowData.Repository),
		views.DefaultRowDataStyle.Render(rowData.TargetName),
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
			w1 := (*workspaceList)[i]
			w2 := (*workspaceList)[j]
			if w1.Info == nil || w2.Info == nil {
				return true
			}
			return w1.Info.Created > w2.Info.Created
		})
		return
	}

	sort.Slice(*workspaceList, func(i, j int) bool {
		w1 := (*workspaceList)[i]
		w2 := (*workspaceList)[j]
		if w1.State == nil || w2.State == nil {
			return true
		}
		return w1.State.Uptime < w2.State.Uptime
	})
}

func getTableRowData(workspace apiclient.WorkspaceDTO, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", "", ""}
	rowData.Name = workspace.Name + views_util.AdditionalPropertyPadding
	rowData.Repository = util.GetRepositorySlugFromUrl(workspace.Repository.Url, specifyGitProviders)
	rowData.Branch = workspace.Repository.Branch

	rowData.TargetName = workspace.TargetName + views_util.AdditionalPropertyPadding

	if workspace.Info != nil {
		rowData.Created = util.FormatTimestamp(workspace.Info.Created)
	}
	if workspace.State != nil && workspace.State.Uptime > 0 {
		rowData.Status = util.FormatUptime(workspace.State.Uptime)
	}
	return &rowData
}
