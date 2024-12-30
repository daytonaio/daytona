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

type rowData struct {
	Id         string
	State      string
	PrebuildId string
	CreatedAt  string
	UpdatedAt  string
}

func ListBuilds(buildList []apiclient.Build, apiServerConfig *apiclient.ServerConfig) {
	if len(buildList) == 0 {
		views_util.NotifyEmptyBuildList(true)
		return
	}

	SortBuilds(&buildList)

	data := [][]string{}

	for _, b := range buildList {
		data = append(data, getRowFromRowData(b))
	}

	table := views_util.GetTableView(data, []string{
		"ID", "State", "Prebuild ID", "Created", "Updated",
	}, nil, func() {
		renderUnstyledList(buildList, apiServerConfig)
	})

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

func getRowFromRowData(build apiclient.Build) []string {
	var data rowData

	data.Id = build.Id + views_util.AdditionalPropertyPadding
	data.State = string(build.State)
	data.PrebuildId = build.PrebuildId
	if data.PrebuildId == "" {
		data.PrebuildId = "/"
	}
	data.CreatedAt = util.FormatTimestamp(build.CreatedAt)
	data.UpdatedAt = util.FormatTimestamp(build.UpdatedAt)

	return []string{
		views.NameStyle.Render(data.Id),
		views.DefaultRowDataStyle.Render(data.State),
		views.DefaultRowDataStyle.Render(data.PrebuildId),
		views.DefaultRowDataStyle.Render(data.CreatedAt),
		views.DefaultRowDataStyle.Render(data.UpdatedAt),
	}
}
