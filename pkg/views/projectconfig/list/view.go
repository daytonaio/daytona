// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/projectconfig/info"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type RowData struct {
	Name       string
	Repository string
	Build      string
	Prebuilds  string
	IsDefault  string
}

func ListProjectConfigs(projectConfigList []apiclient.ProjectConfig, apiServerConfig *apiclient.ServerConfig, specifyGitProviders bool) {
	data := [][]string{}

	for _, pc := range projectConfigList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(pc, apiServerConfig, specifyGitProviders)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table, success := views_util.GetTableView(data, []string{
		"Name", "Repository", "Build", "Prebuild rules", "Default",
	}, nil)

	if !success {
		renderUnstyledList(projectConfigList, apiServerConfig)
		return
	}

	fmt.Println(table)
}

func renderUnstyledList(projectConfigList []apiclient.ProjectConfig, apiServerConfig *apiclient.ServerConfig) {
	for _, pc := range projectConfigList {
		info.Render(&pc, apiServerConfig, true)

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
		views.DefaultRowDataStyle.Render(rowData.Prebuilds),
		isDefault,
	}

	return row
}

func getTableRowData(projectConfig apiclient.ProjectConfig, apiServerConfig *apiclient.ServerConfig, specifyGitProviders bool) *RowData {
	rowData := RowData{"", "", "", "", ""}

	rowData.Name = projectConfig.Name + views_util.AdditionalPropertyPadding
	rowData.Repository = util.GetRepositorySlugFromUrl(projectConfig.RepositoryUrl, specifyGitProviders)
	rowData.Prebuilds = "None"
	rowData.IsDefault = ""

	projectDefaults := &views_util.ProjectConfigDefaults{
		Image:     &apiServerConfig.DefaultProjectImage,
		ImageUser: &apiServerConfig.DefaultProjectUser,
	}

	createProjectDto := apiclient.CreateProjectDTO{
		BuildConfig: projectConfig.BuildConfig,
	}

	_, rowData.Build = views_util.GetProjectBuildChoice(createProjectDto, projectDefaults)

	if projectConfig.Default {
		rowData.IsDefault = "1"
	}

	if len(projectConfig.Prebuilds) > 0 {
		rowData.Prebuilds = fmt.Sprintf("%d", len(projectConfig.Prebuilds))
	}

	return &rowData
}
