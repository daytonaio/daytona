// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	info_view "github.com/daytonaio/daytona/pkg/views/target/info"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type RowData struct {
	Name           string
	Provider       string
	Default        bool
	WorkspaceCount string
	Options        string
}

func ListTargets(targetList []apiclient.TargetDTO, verbose bool, activeProfileName string) {
	if len(targetList) == 0 {
		views_util.NotifyEmptyTargetList(true)
		return
	}

	SortTargets(&targetList)

	headers := []string{"Target", "Provider", "Default", "# Workspaces", "Options"}

	data := util.ArrayMap(targetList, func(target apiclient.TargetDTO) []string {
		rowData := RowData{
			Name:           target.Name,
			Provider:       target.ProviderInfo.Name,
			Default:        target.Default,
			WorkspaceCount: fmt.Sprintf("%d", target.WorkspaceCount),
			Options:        target.Options,
		}

		return getRowFromRowData(rowData)
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
	var isDefault string

	if rowData.Default {
		isDefault = views.ActiveStyle.Render("Yes")
	} else {
		isDefault = views.InactiveStyle.Render("/")
	}

	return []string{
		views.NameStyle.Render(rowData.Name),
		views.DefaultRowDataStyle.Render(rowData.Provider),
		isDefault,
		views.DefaultRowDataStyle.Render(rowData.WorkspaceCount),
		views.DefaultRowDataStyle.Render(rowData.Options),
	}
}

func SortTargets(targetList *[]apiclient.TargetDTO) {
	sort.Slice(*targetList, func(i, j int) bool {
		return (*targetList)[i].Default && !(*targetList)[j].Default
	})
}
