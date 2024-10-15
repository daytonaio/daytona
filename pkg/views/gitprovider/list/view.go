// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/gitprovider"
	"github.com/daytonaio/daytona/pkg/views/gitprovider/info"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type rowData struct {
	Name       string
	Alias      string
	Username   string
	BaseApiUrl string
}

func ListGitProviders(gitProviderViewList []gitprovider.GitProviderView) {
	var showBaseApiUrlColumn bool
	headers := []string{"Name", "Alias", "Username", "Base API URL"}

	for _, gp := range gitProviderViewList {
		if gp.BaseApiUrl != "" {
			showBaseApiUrlColumn = true
			break
		}
	}

	data := [][]string{}

	for _, b := range gitProviderViewList {
		data = append(data, getRowFromRowData(b))
	}

	if !showBaseApiUrlColumn {
		headers = headers[:len(headers)-1]
		for value := range data {
			data[value] = data[value][:len(data[value])-1]
		}
	}

	table := views_util.GetTableView(data, headers, nil, func() {
		renderUnstyledList(gitProviderViewList)
	})

	fmt.Println(table)
}

func renderUnstyledList(gitProviderViewList []gitprovider.GitProviderView) {
	for _, b := range gitProviderViewList {
		info.Render(&b, true)

		if b.Id != gitProviderViewList[len(gitProviderViewList)-1].Id {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}
	}
}

func getRowFromRowData(build gitprovider.GitProviderView) []string {
	var data rowData

	data.Name = build.Name + views_util.AdditionalPropertyPadding
	data.Alias = build.Alias
	data.Username = build.Username
	data.BaseApiUrl = build.BaseApiUrl

	return []string{
		views.NameStyle.Render(data.Name),
		views.DefaultRowDataStyle.Render(data.Alias),
		views.DefaultRowDataStyle.Render(data.Username),
		views.DefaultRowDataStyle.Render(data.BaseApiUrl),
	}
}
