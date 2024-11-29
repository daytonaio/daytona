// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspacetemplate/info"
)

type rowData struct {
	Name       string
	Repository string
	Build      string
	Prebuilds  string
	IsDefault  string
}

func ListWorkspaceTemplates(workspaceTemplateList []apiclient.WorkspaceTemplate, apiServerConfig *apiclient.ServerConfig, specifyGitProviders bool) {
	if len(workspaceTemplateList) == 0 {
		views_util.NotifyEmptyWorkspaceTemplateList(true)
		return
	}

	data := [][]string{}

	for _, wt := range workspaceTemplateList {
		data = append(data, getRowFromData(wt, apiServerConfig, specifyGitProviders))
	}

	table := views_util.GetTableView(data, []string{
		"Name", "Repository", "Build", "Prebuild rules", "Default",
	}, nil, func() {
		renderUnstyledList(workspaceTemplateList, apiServerConfig)
	})

	fmt.Println(table)
}

func renderUnstyledList(workspaceTemplateList []apiclient.WorkspaceTemplate, apiServerConfig *apiclient.ServerConfig) {
	for _, wt := range workspaceTemplateList {
		info.Render(&wt, apiServerConfig, true)

		if wt.Name != workspaceTemplateList[len(workspaceTemplateList)-1].Name {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}
	}
}

func getRowFromData(workspaceTemplate apiclient.WorkspaceTemplate, apiServerConfig *apiclient.ServerConfig, specifyGitProviders bool) []string {
	var isDefault string
	var data rowData

	data.Name = workspaceTemplate.Name + views_util.AdditionalPropertyPadding
	data.Repository = util.GetRepositorySlugFromUrl(workspaceTemplate.RepositoryUrl, specifyGitProviders)
	data.Prebuilds = "None"

	workspaceDefaults := &views_util.WorkspaceTemplateDefaults{
		Image:     &apiServerConfig.DefaultWorkspaceImage,
		ImageUser: &apiServerConfig.DefaultWorkspaceUser,
	}

	createWorkspaceDto := apiclient.CreateWorkspaceDTO{
		BuildConfig: workspaceTemplate.BuildConfig,
	}

	_, data.Build = views_util.GetWorkspaceBuildChoice(createWorkspaceDto, workspaceDefaults)

	if workspaceTemplate.Default {
		isDefault = views.ActiveStyle.Render("Yes")
	} else {
		isDefault = views.InactiveStyle.Render("/")
	}

	if len(workspaceTemplate.Prebuilds) > 0 {
		data.Prebuilds = fmt.Sprintf("%d", len(workspaceTemplate.Prebuilds))
	}

	return []string{
		views.NameStyle.Render(data.Name),
		views.DefaultRowDataStyle.Render(data.Repository),
		views.DefaultRowDataStyle.Render(data.Build),
		views.DefaultRowDataStyle.Render(data.Prebuilds),
		isDefault,
	}
}
