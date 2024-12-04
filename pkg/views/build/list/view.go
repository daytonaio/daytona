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
	Status     string
	PrebuildId string
	CreatedAt  string
	UpdatedAt  string
}

func ListBuilds(buildList []apiclient.BuildDTO, apiServerConfig *apiclient.ServerConfig) {
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

func SortBuilds(buildList *[]apiclient.BuildDTO) {
	sort.Slice(*buildList, func(i, j int) bool {
		pi, pj := views_util.GetStateSortPriorities((*buildList)[i].State.Name, (*buildList)[j].State.Name)
		if pi != pj {
			return pi < pj
		}

		// If two builds have the same state priority, compare the UpdatedAt property
		return (*buildList)[i].State.UpdatedAt > (*buildList)[j].State.UpdatedAt
	})
}

func renderUnstyledList(buildList []apiclient.BuildDTO, apiServerConfig *apiclient.ServerConfig) {
	for _, b := range buildList {
		info.Render(&b, apiServerConfig, true)

		if b.Id != buildList[len(buildList)-1].Id {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}
	}
}

func getRowFromRowData(build apiclient.BuildDTO) []string {
	var data rowData

	data.Id = build.Id + views_util.AdditionalPropertyPadding
	data.Status = views.GetStateLabel(build.State.Name)
	data.PrebuildId = build.PrebuildId
	if data.PrebuildId == "" {
		data.PrebuildId = "/"
	}
	data.CreatedAt = util.FormatTimestamp(build.CreatedAt)
	data.UpdatedAt = util.FormatTimestamp(build.UpdatedAt)

	return []string{
		views.NameStyle.Render(data.Id),
		views.DefaultRowDataStyle.Render(data.Status),
		views.DefaultRowDataStyle.Render(data.PrebuildId),
		views.DefaultRowDataStyle.Render(data.CreatedAt),
		views.DefaultRowDataStyle.Render(data.UpdatedAt),
	}
}
