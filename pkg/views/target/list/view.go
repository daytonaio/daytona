// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	info_view "github.com/daytonaio/daytona/pkg/views/target/info"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type RowData struct {
	Name         string
	TargetConfig string
}

func ListTargets(targetList []apiclient.TargetDTO, verbose bool, activeProfileName string) {
	if len(targetList) == 0 {
		views_util.NotifyEmptyTargetList(true)
		return
	}

	headers := []string{"Target", "Target Config"}

	data := util.ArrayMap(targetList, func(target apiclient.TargetDTO) []string {
		return getRowFromRowData(RowData{Name: target.Name, TargetConfig: target.TargetConfig})
	})

	footer := lipgloss.NewStyle().Foreground(views.LightGray).Render(views.GetListFooter(activeProfileName, &views.Padding{}))

	table := views_util.GetTableView(data, headers, &footer, func() {
		renderUnstyledList(targetList)
	})

	fmt.Println(table)
}

func renderUnstyledList(targetList []apiclient.TargetDTO) {
	for _, target := range targetList {
		info_view.Render(&target, true)

		if target.Id != targetList[len(targetList)-1].Id {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}
	}
}

func getRowFromRowData(rowData RowData) []string {
	return []string{
		views.NameStyle.Render(rowData.Name),
		views.DefaultRowDataStyle.Render(rowData.TargetConfig),
	}
}
