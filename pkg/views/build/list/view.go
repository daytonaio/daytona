// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"sort"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/build/info"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type RowData struct {
	Id         string
	State      string
	PrebuildId string
	CreatedAt  string
	UpdatedAt  string
}

func ListBuilds(buildList []apiclient.Build, apiServerConfig *apiclient.ServerConfig) {
	SortBuilds(&buildList)

	data := [][]string{}

	for _, pc := range buildList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(pc)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table, success := views_util.GetTableView(data, []string{
		"ID", "State", "Prebuild ID", "Created", "Updated",
	}, nil)

	if !success {
		renderUnstyledList(buildList, apiServerConfig)
		return
	}

	fmt.Println(table)
}

func SortBuilds(buildList *[]apiclient.Build) {
	sort.Slice(*buildList, func(i, j int) bool {
		b1 := (*buildList)[i]
		b2 := (*buildList)[j]
		return b1.UpdatedAt > b2.UpdatedAt
	})
}

func renderUnstyledList(buildList []apiclient.Build, apiServerConfig *apiclient.ServerConfig) {
	for _, b := range buildList {
		info.Render(&b, apiServerConfig, true)

		if b.Id != buildList[len(buildList)-1].Id {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}
	}
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.Id),
		views.DefaultRowDataStyle.Render(rowData.State),
		views.DefaultRowDataStyle.Render(rowData.PrebuildId),
		views.DefaultRowDataStyle.Render(rowData.CreatedAt),
		views.DefaultRowDataStyle.Render(rowData.UpdatedAt),
	}

	return row
}

func getTableRowData(build apiclient.Build) *RowData {
	rowData := RowData{"", "", "", "", ""}

	rowData.Id = build.Id + views_util.AdditionalPropertyPadding
	rowData.State = string(build.State)
	rowData.PrebuildId = build.PrebuildId
	if rowData.PrebuildId == "" {
		rowData.PrebuildId = "/"
	}
	rowData.CreatedAt = util.FormatTimestamp(build.CreatedAt)
	rowData.UpdatedAt = util.FormatTimestamp(build.UpdatedAt)

	return &rowData
}
