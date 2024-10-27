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
	info_view "github.com/daytonaio/daytona/pkg/views/target/info"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type RowData struct {
	Name         string
	Repository   string
	TargetConfig string
	Status       string
	Created      string
	Branch       string
}

func ListTargets(targetList []apiclient.TargetDTO, specifyGitProviders bool, verbose bool, activeProfileName string) {
	SortTargets(&targetList, verbose)

	headers := []string{"Target", "Repository", "Target Config", "Status", "Created", "Branch"}

	data := [][]string{}

	for _, target := range targetList {
		var rowData *RowData
		var row []string

		if len(target.Workspaces) == 1 {
			rowData = getTargetTableRowData(target, specifyGitProviders)
			row = getRowFromRowData(*rowData, false)
			data = append(data, row)
		} else {
			row = getRowFromRowData(RowData{Name: target.Name}, true)
			data = append(data, row)
			for _, workspace := range target.Workspaces {
				rowData = getWorkspaceTableRowData(target, workspace, specifyGitProviders)
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
		renderUnstyledList(targetList)
	})

	fmt.Println(table)
}

func renderUnstyledList(targetList []apiclient.TargetDTO) {
	for _, target := range targetList {
		info_view.Render(&target, "", true)

		if target.Id != targetList[len(targetList)-1].Id {
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
		views.DefaultRowDataStyle.Render(rowData.TargetConfig),
		state,
		views.DefaultRowDataStyle.Render(rowData.Created),
		views.DefaultRowDataStyle.Render(views.GetBranchNameLabel(rowData.Branch)),
	}

	if rowData.Status != "" {
		row[3] = fmt.Sprintf("%s %s", state, views.DefaultRowDataStyle.Render(fmt.Sprintf("(%s)", rowData.Status)))
	}

	return row
}

func SortTargets(targetList *[]apiclient.TargetDTO, verbose bool) {
	if verbose {
		sort.Slice(*targetList, func(i, j int) bool {
			t1 := (*targetList)[i]
			t2 := (*targetList)[j]
			if t1.Info == nil || t2.Info == nil || t1.Info.Workspaces == nil || t2.Info.Workspaces == nil {
				return true
			}
			if len(t1.Info.Workspaces) == 0 || len(t2.Info.Workspaces) == 0 {
				return true
			}
			return t1.Info.Workspaces[0].Created > t2.Info.Workspaces[0].Created
		})
		return
	}

	sort.Slice(*targetList, func(i, j int) bool {
		t1 := (*targetList)[i]
		t2 := (*targetList)[j]
		if len(t1.Workspaces) == 0 || len(t2.Workspaces) == 0 || t1.Workspaces[0].State == nil || t2.Workspaces[0].State == nil {
			return true
		}
		return t1.Workspaces[0].State.Uptime < t2.Workspaces[0].State.Uptime
	})

}

func getTargetTableRowData(target apiclient.TargetDTO, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", "", ""}
	rowData.Name = target.Name + views_util.AdditionalPropertyPadding
	if len(target.Workspaces) > 0 {
		rowData.Repository = util.GetRepositorySlugFromUrl(target.Workspaces[0].Repository.Url, specifyGitProviders)
		rowData.Branch = target.Workspaces[0].Repository.Branch
	}

	rowData.TargetConfig = target.TargetConfig + views_util.AdditionalPropertyPadding

	if target.Info != nil && target.Info.Workspaces != nil && len(target.Info.Workspaces) > 0 {
		rowData.Created = util.FormatTimestamp(target.Info.Workspaces[0].Created)
	}
	if len(target.Workspaces) > 0 && target.Workspaces[0].State != nil && target.Workspaces[0].State.Uptime > 0 {
		rowData.Status = util.FormatUptime(target.Workspaces[0].State.Uptime)
	}
	return &rowData
}

func getWorkspaceTableRowData(targetDTO apiclient.TargetDTO, workspace apiclient.Workspace, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", "", ""}
	rowData.Name = " └ " + workspace.Name

	rowData.Repository = util.GetRepositorySlugFromUrl(workspace.Repository.Url, specifyGitProviders)
	rowData.Branch = workspace.Repository.Branch

	rowData.TargetConfig = workspace.TargetConfig + views_util.AdditionalPropertyPadding

	if workspace.State != nil && workspace.State.Uptime > 0 {
		rowData.Status = util.FormatUptime(workspace.State.Uptime)
	}

	if targetDTO.Info == nil || targetDTO.Info.Workspaces == nil {
		return &rowData
	}

	for _, workspaceInfo := range targetDTO.Info.Workspaces {
		if workspaceInfo.Name == workspace.Name {
			rowData.Created = util.FormatTimestamp(workspaceInfo.Created)
			break
		}
	}

	return &rowData
}
