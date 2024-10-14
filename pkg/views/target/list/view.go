// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"sort"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"
)

type RowData struct {
	Target   string
	Provider string
	Options  string
}

func ListTargets(targetList []apiclient.ProviderTarget) {
	sortTargets(&targetList)

	data := [][]string{}

	for _, target := range targetList {
		var rowData *RowData
		var row []string

		rowData = getRowData(&target)
		if rowData == nil {
			continue
		}
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table, success := util.GetTableView(data, []string{
		"Target", "Provider", "Options",
	}, nil)

	if !success {
		renderUnstyledList(targetList)
		return
	}

	fmt.Println(table)
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.Target),
		views.DefaultRowDataStyle.Render(rowData.Provider),
		views.DefaultRowDataStyle.Render(rowData.Options),
	}

	return row
}

func getRowData(target *apiclient.ProviderTarget) *RowData {
	rowData := RowData{"", "", ""}

	rowData.Target = target.Name
	rowData.Provider = target.ProviderInfo.Name
	rowData.Options = target.Options

	return &rowData
}

func sortTargets(targets *[]apiclient.ProviderTarget) {
	sort.Slice(*targets, func(i, j int) bool {
		t1 := (*targets)[i]
		t2 := (*targets)[j]
		return t1.ProviderInfo.Name < t2.ProviderInfo.Name
	})
}

func renderUnstyledList(targetList []apiclient.ProviderTarget) {
	output := "\n"

	for _, target := range targetList {
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Target Name: "), target.Name) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Target Provider: "), target.ProviderInfo.Name) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Target Options: "), target.Options) + "\n"

		if target.Name != targetList[len(targetList)-1].Name {
			output += views.SeparatorString + "\n\n"
		}
	}

	fmt.Println(output)
}
