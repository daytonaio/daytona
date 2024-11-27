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
	Uptime     string
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
	}

	footer := lipgloss.NewStyle().Foreground(views.LightGray).Render(views.GetListFooter(activeProfileName, &views.Padding{}))

	table := views_util.GetTableView(data, headers, &footer, func() {
		renderUnstyledList(workspaceList)
	})

	fmt.Println(table)
}

func SortWorkspaces(workspaceList *[]apiclient.WorkspaceDTO, verbose bool) {
	sort.Slice(*workspaceList, func(i, j int) bool {
		pi, ok := views.ResourceListStatePriorities[(*workspaceList)[i].State.Name]
		if !ok {
			pi = 99
		}
		pj, ok2 := views.ResourceListStatePriorities[(*workspaceList)[j].State.Name]
		if !ok2 {
			pj = 99
		}

		if pi != pj {
			return pi < pj
		}

		// If two workspaces have the same state priority, compare the UpdatedAt property
		return (*workspaceList)[i].State.UpdatedAt > (*workspaceList)[j].State.UpdatedAt
	})
}

func getTableRowData(workspace apiclient.WorkspaceDTO, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", "", "", ""}
	rowData.Name = workspace.Name + views_util.AdditionalPropertyPadding
	rowData.Repository = util.GetRepositorySlugFromUrl(workspace.Repository.Url, specifyGitProviders)
	rowData.Branch = workspace.Repository.Branch
	rowData.Status = views.GetStateLabel(workspace.State.Name)

	rowData.TargetName = workspace.Target.Name + views_util.AdditionalPropertyPadding

	if workspace.Info != nil {
		rowData.Created = util.FormatTimestamp(workspace.Info.Created)
	}

	if workspace.Metadata != nil {
		views_util.CheckAndAppendTimeLabel(&rowData.Status, workspace.State, workspace.Metadata.Uptime)
	}

	return &rowData
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
	if isMultiWorkspaceAccordion {
		return []string{rowData.Name, "", "", "", "", ""}
	}

	row := []string{
		views.NameStyle.Render(rowData.Name),
		views.DefaultRowDataStyle.Render(rowData.Repository),
		views.DefaultRowDataStyle.Render(rowData.TargetName),
		rowData.Status,
		views.DefaultRowDataStyle.Render(rowData.Created),
		views.DefaultRowDataStyle.Render(views.GetBranchNameLabel(rowData.Branch)),
	}

	return row
}
