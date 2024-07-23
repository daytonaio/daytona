// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"golang.org/x/term"
)

type RowData struct {
	Name       string
	Repository string
	Build      string
	IsDefault  string
}

func ListProjectConfigs(projectConfigList []apiclient.ProjectConfig, apiServerConfig *apiclient.ServerConfig, specifyGitProviders bool) {
	re := lipgloss.NewRenderer(os.Stdout)

	headers := []string{"Name", "Repository", "Build", "Default"}

	data := [][]string{}

	for _, pc := range projectConfigList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(pc, apiServerConfig, specifyGitProviders)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(data)
		return
	}

	breakpointWidth := views.GetContainerBreakpointWidth(terminalWidth)

	minWidth := views_util.GetTableMinimumWidth(data)

	if breakpointWidth == 0 || minWidth > breakpointWidth {
		renderUnstyledList(projectConfigList)
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

func renderUnstyledList(projectConfigList []apiclient.ProjectConfig) {
	for _, pc := range projectConfigList {
		// render info
		// info_view.Render(&workspace, "", true)

		if pc.Name != projectConfigList[len(projectConfigList)-1].Name {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}

	}
}

func getRowFromRowData(rowData RowData) []string {
	var isDefault string

	if rowData.IsDefault == "" {
		isDefault = views.InactiveStyle.Render("/")
	} else {
		isDefault = views.ActiveStyle.Render("Yes")
	}

	row := []string{
		views.NameStyle.Render(rowData.Name),
		views.DefaultRowDataStyle.Render(rowData.Repository),
		views.DefaultRowDataStyle.Render(rowData.Build),
		isDefault,
	}

	return row
}

func getTableRowData(projectConfig apiclient.ProjectConfig, apiServerConfig *apiclient.ServerConfig, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", ""}

	rowData.Name = *projectConfig.Name + views_util.AdditionalPropertyPadding
	rowData.Repository = util.GetRepositorySlugFromUrl(*projectConfig.Repository.Url, specifyGitProviders)
	rowData.IsDefault = ""

	projectDefaults := &create.ProjectDefaults{
		Image:     apiServerConfig.DefaultProjectImage,
		ImageUser: apiServerConfig.DefaultProjectUser,
	}

	createProjectConfigDTO := apiclient.CreateProjectConfigDTO{
		Build: projectConfig.Build,
	}

	_, rowData.Build = create.GetProjectBuildChoice(createProjectConfigDTO, projectDefaults)

	if *projectConfig.Default {
		rowData.IsDefault = "1"
	}

	return &rowData
}
