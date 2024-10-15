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

type rowData struct {
	Name       string
	Repository string
	Build      string
	Prebuilds  string
	IsDefault  string
}

func ListProjectConfigs(projectConfigList []apiclient.ProjectConfig, apiServerConfig *apiclient.ServerConfig, specifyGitProviders bool) {
	data := [][]string{}

	for _, pc := range projectConfigList {
		data = append(data, getRowFromData(pc, apiServerConfig, specifyGitProviders))
	}

	table := views_util.GetTableView(data, []string{
		"Name", "Repository", "Build", "Prebuild rules", "Default",
	}, nil, renderUnstyledList, projectConfigList, apiServerConfig)

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

func getRowFromData(projectConfig apiclient.ProjectConfig, apiServerConfig *apiclient.ServerConfig, specifyGitProviders bool) []string {
	var isDefault string
	var data rowData

	data.Name = projectConfig.Name + views_util.AdditionalPropertyPadding
	data.Repository = util.GetRepositorySlugFromUrl(projectConfig.RepositoryUrl, specifyGitProviders)
	data.Prebuilds = "None"
	data.IsDefault = ""

	projectDefaults := &views_util.ProjectConfigDefaults{
		Image:     &apiServerConfig.DefaultProjectImage,
		ImageUser: &apiServerConfig.DefaultProjectUser,
	}

	createProjectDto := apiclient.CreateProjectDTO{
		BuildConfig: projectConfig.BuildConfig,
	}

	_, data.Build = views_util.GetProjectBuildChoice(createProjectDto, projectDefaults)

	if projectConfig.Default {
		data.IsDefault = "1"
	}

	if len(projectConfig.Prebuilds) > 0 {
		data.Prebuilds = fmt.Sprintf("%d", len(projectConfig.Prebuilds))
	}

	if data.IsDefault == "" {
		isDefault = views.InactiveStyle.Render("/")
	} else {
		isDefault = views.ActiveStyle.Render("Yes")
	}

	return []string{
		views.NameStyle.Render(data.Name),
		views.DefaultRowDataStyle.Render(data.Repository),
		views.DefaultRowDataStyle.Render(data.Build),
		views.DefaultRowDataStyle.Render(data.Prebuilds),
		isDefault,
	}
}
