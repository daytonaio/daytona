// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"
)

type RowData struct {
	Label   string
	Name    string
	Version string
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.Label),
		views.DefaultRowDataStyle.Render(rowData.Name),
		views.DefaultRowDataStyle.Render(rowData.Version),
	}

	return row
}

func getRowData(provider *apiclient.Provider) *RowData {
	rowData := RowData{"", "", ""}

	if provider.Label != nil {
		rowData.Label = *provider.Label
	} else {
		rowData.Label = provider.Name
	}
	rowData.Name = provider.Name
	rowData.Version = provider.Version

	return &rowData
}

func List(providerList []apiclient.Provider) {
	data := [][]string{}

	for _, provider := range providerList {
		var rowData *RowData
		var row []string

		rowData = getRowData(&provider)
		if rowData == nil {
			continue
		}
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table, success := util.GetTableView(data, []string{
		"Provider", "Name", "Version",
	}, nil)

	if !success {
		renderUnstyledList(providerList)
		return
	}

	fmt.Println(table)
}

func renderUnstyledList(providerList []apiclient.Provider) {
	output := "\n"

	for _, provider := range providerList {
		if provider.Label != nil {
			output += fmt.Sprintf("%s %s", views.GetPropertyKey("Provider: "), *provider.Label) + "\n\n"
		}
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Name: "), provider.Name) + "\n\n"
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Version: "), provider.Version) + "\n"

		if provider.Name != providerList[len(providerList)-1].Name {
			output += views.SeparatorString + "\n\n"
		}
	}

	fmt.Println(output)
}
