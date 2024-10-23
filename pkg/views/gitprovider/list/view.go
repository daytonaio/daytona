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
	Name          string
	Alias         string
	Username      string
	BaseApiUrl    string
	SigningMethod string
}

func ListGitProviders(gitProviderViewList []gitprovider.GitProviderView) {
	var showBaseApiUrlColumn bool
	var showSigningMethodColumn bool
	headers := []string{"Name", "Alias", "Username", "Base API URL", "Signing Method"}

	for _, gp := range gitProviderViewList {
		if gp.BaseApiUrl != "" {
			showBaseApiUrlColumn = true
		}
		if gp.SigningMethod != "" {
			showSigningMethodColumn = true
		}
	}

	data := [][]string{}

	for _, b := range gitProviderViewList {
		data = append(data, getRowFromRowData(b))
	}

	if !showBaseApiUrlColumn {
		headers = removeHeader(headers, "Base API URL")
		for i := range data {
			data[i] = removeColumn(data[i], 3)
		}
	}
	if !showSigningMethodColumn {
		headers = removeHeader(headers, "Signing Method")
		for i := range data {
			data[i] = removeColumn(data[i], 4)
		}
	}

	table := views_util.GetTableView(data, headers, nil, func() {
		renderUnstyledList(gitProviderViewList)
	})

	fmt.Println(table)
}

func removeHeader(headers []string, headerToRemove string) []string {
	for i, header := range headers {
		if header == headerToRemove {
			return append(headers[:i], headers[i+1:]...)
		}
	}
	return headers
}

func removeColumn(data []string, index int) []string {
	if index < 0 || index >= len(data) {
		return data
	}
	return append(data[:index], data[index+1:]...)
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
	data.SigningMethod = build.SigningMethod

	return []string{
		views.NameStyle.Render(data.Name),
		views.DefaultRowDataStyle.Render(data.Alias),
		views.DefaultRowDataStyle.Render(data.Username),
		views.DefaultRowDataStyle.Render(data.BaseApiUrl),
		views.DefaultRowDataStyle.Render(data.SigningMethod),
	}
}
