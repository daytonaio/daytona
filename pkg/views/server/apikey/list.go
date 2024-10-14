// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"
)

type RowData struct {
	Name string
	Type string
}

func ListApiKeys(apiKeyList []apiclient.ApiKey) {
	data := [][]string{}

	for _, apiKey := range apiKeyList {
		var rowData *RowData
		var row []string

		rowData = getRowData(&apiKey)
		if rowData == nil {
			continue
		}
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table, success := util.GetTableView(data, []string{
		"Name", "Type",
	}, nil)

	if !success {
		renderUnstyledList(apiKeyList)
		return
	}

	fmt.Println(table)
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.Name),
		views.DefaultRowDataStyle.Render(rowData.Type),
	}

	return row
}

func getRowData(apiKey *apiclient.ApiKey) *RowData {
	rowData := RowData{"", ""}

	rowData.Name = apiKey.Name
	rowData.Type = string(apiKey.Type)

	return &rowData
}

func renderUnstyledList(apiKeyList []apiclient.ApiKey) {
	output := "\n"

	for _, apiKey := range apiKeyList {
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("API Key Name: "), apiKey.Name) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("API Key Type: "), apiKey.Type) + "\n\n"

		if apiKey.Name != apiKeyList[len(apiKeyList)-1].Name {
			output += views.SeparatorString + "\n\n"
		}
	}

	fmt.Println(output)
}
