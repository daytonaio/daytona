// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspaceconfig/info"
)

type rowData struct {
	Name       string
	Repository string
	Build      string
	Prebuilds  string
	IsDefault  string
}

func ListWorkspaceConfigs(workspaceConfigList []apiclient.WorkspaceConfig, apiServerConfig *apiclient.ServerConfig, specifyGitProviders bool) {
	if len(workspaceConfigList) == 0 {
		views_util.NotifyEmptyWorkspaceConfigList(true)
		return
	}

	data := [][]string{}

	for _, wc := range workspaceConfigList {
		data = append(data, getRowFromData(wc, apiServerConfig, specifyGitProviders))
	}

	table := views_util.GetTableView(data, []string{
		"Name", "Repository", "Build", "Prebuild rules", "Default",
	}, nil, func() {
		renderUnstyledList(workspaceConfigList, apiServerConfig)
	})

	fmt.Println(table)
}

func renderUnstyledList(workspaceConfigList []apiclient.WorkspaceConfig, apiServerConfig *apiclient.ServerConfig) {
	for _, wc := range workspaceConfigList {
		info.Render(&wc, apiServerConfig, true)

		if wc.Name != workspaceConfigList[len(workspaceConfigList)-1].Name {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}
	}
}

func getRowFromData(workspaceConfig apiclient.WorkspaceConfig, apiServerConfig *apiclient.ServerConfig, specifyGitProviders bool) []string {
	var isDefault string
	var data rowData

	data.Name = workspaceConfig.Name + views_util.AdditionalPropertyPadding
	data.Repository = util.GetRepositorySlugFromUrl(workspaceConfig.RepositoryUrl, specifyGitProviders)
	data.Prebuilds = "None"

	workspaceDefaults := &views_util.WorkspaceConfigDefaults{
		Image:     &apiServerConfig.DefaultWorkspaceImage,
		ImageUser: &apiServerConfig.DefaultWorkspaceUser,
	}

	createWorkspaceDto := apiclient.CreateWorkspaceDTO{
		BuildConfig: workspaceConfig.BuildConfig,
	}

	_, data.Build = views_util.GetWorkspaceBuildChoice(createWorkspaceDto, workspaceDefaults)

	if workspaceConfig.Default {
		isDefault = views.ActiveStyle.Render("Yes")
	} else {
		isDefault = views.InactiveStyle.Render("/")
	}

	if len(workspaceConfig.Prebuilds) > 0 {
		data.Prebuilds = fmt.Sprintf("%d", len(workspaceConfig.Prebuilds))
	}

	return []string{
		views.NameStyle.Render(data.Name),
		views.DefaultRowDataStyle.Render(data.Repository),
		views.DefaultRowDataStyle.Render(data.Build),
		views.DefaultRowDataStyle.Render(data.Prebuilds),
		isDefault,
	}
}
