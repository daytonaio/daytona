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

		if len(target.Projects) == 1 {
			rowData = getTargetTableRowData(target, specifyGitProviders)
			row = getRowFromRowData(*rowData, false)
			data = append(data, row)
		} else {
			row = getRowFromRowData(RowData{Name: target.Name}, true)
			data = append(data, row)
			for _, project := range target.Projects {
				rowData = getProjectTableRowData(target, project, specifyGitProviders)
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
			if t1.Info == nil || t2.Info == nil || t1.Info.Projects == nil || t2.Info.Projects == nil {
				return true
			}
			if len(t1.Info.Projects) == 0 || len(t2.Info.Projects) == 0 {
				return true
			}
			return t1.Info.Projects[0].Created > t2.Info.Projects[0].Created
		})
		return
	}

	sort.Slice(*targetList, func(i, j int) bool {
		t1 := (*targetList)[i]
		t2 := (*targetList)[j]
		if len(t1.Projects) == 0 || len(t2.Projects) == 0 || t1.Projects[0].State == nil || t2.Projects[0].State == nil {
			return true
		}
		return t1.Projects[0].State.Uptime < t2.Projects[0].State.Uptime
	})

}

func getTargetTableRowData(target apiclient.TargetDTO, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", "", ""}
	rowData.Name = target.Name + views_util.AdditionalPropertyPadding
	if len(target.Projects) > 0 {
		rowData.Repository = util.GetRepositorySlugFromUrl(target.Projects[0].Repository.Url, specifyGitProviders)
		rowData.Branch = target.Projects[0].Repository.Branch
	}

	rowData.TargetConfig = target.TargetConfig + views_util.AdditionalPropertyPadding

	if target.Info != nil && target.Info.Projects != nil && len(target.Info.Projects) > 0 {
		rowData.Created = util.FormatTimestamp(target.Info.Projects[0].Created)
	}
	if len(target.Projects) > 0 && target.Projects[0].State != nil && target.Projects[0].State.Uptime > 0 {
		rowData.Status = util.FormatUptime(target.Projects[0].State.Uptime)
	}
	return &rowData
}

func getProjectTableRowData(targetDTO apiclient.TargetDTO, project apiclient.Project, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", "", ""}
	rowData.Name = " └ " + project.Name

	rowData.Repository = util.GetRepositorySlugFromUrl(project.Repository.Url, specifyGitProviders)
	rowData.Branch = project.Repository.Branch

	rowData.TargetConfig = project.TargetConfig + views_util.AdditionalPropertyPadding

	if project.State != nil && project.State.Uptime > 0 {
		rowData.Status = util.FormatUptime(project.State.Uptime)
	}

	if targetDTO.Info == nil || targetDTO.Info.Projects == nil {
		return &rowData
	}

	for _, projectInfo := range targetDTO.Info.Projects {
		if projectInfo.Name == project.Name {
			rowData.Created = util.FormatTimestamp(projectInfo.Created)
			break
		}
	}

	return &rowData
}
