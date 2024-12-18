// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/server/runner/info"
	"github.com/daytonaio/daytona/pkg/views/util"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type rowData struct {
	Name  string
	Id    string
	State string
}

func ListRunners(runnerList []apiclient.RunnerDTO) {
	if len(runnerList) == 0 {
		views_util.NotifyEmptyRunnerList(true)
		return
	}

	data := [][]string{}

	for _, p := range runnerList {
		data = append(data, getRowFromData(p))
	}

	table := util.GetTableView(data, []string{
		"Name", "ID", "State",
	}, nil, func() {
		renderUnstyledList(runnerList)
	})

	fmt.Println(table)
}

func renderUnstyledList(runnerList []apiclient.RunnerDTO) {
	for _, p := range runnerList {
		info.Render(&p, true)

		if p.Id != runnerList[len(runnerList)-1].Id {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}
	}
}

func getRowFromData(runner apiclient.RunnerDTO) []string {
	var data rowData

	data.Name = runner.Name + views_util.AdditionalPropertyPadding
	data.Id = runner.Id
	data.State = runner.State

	return []string{
		views.NameStyle.Render(data.Name),
		views.DefaultRowDataStyle.Render(data.Id),
		views.DefaultRowDataStyle.Render(data.State),
	}
}
