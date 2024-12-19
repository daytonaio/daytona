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
	WorkspaceCount string
	Default        bool
	Status         string
	Options        string
	Uptime         string
}

func ListTargets(targetList []apiclient.TargetDTO, activeProfileName string) {
	if len(targetList) == 0 {
		views_util.NotifyEmptyTargetList(true)
		return
	}

	SortTargets(&targetList)

	headers := []string{"Target", "Options", "# Workspaces", "Default", "Status"}

	data := util.ArrayMap(targetList, func(target apiclient.TargetDTO) []string {
		provider := target.TargetConfig.ProviderInfo.Name
		if target.TargetConfig.ProviderInfo.Label != nil {
			provider = *target.TargetConfig.ProviderInfo.Label
		}

		rowData := RowData{
			Name:           target.Name,
			Provider:       provider,
			Options:        target.TargetConfig.Options,
			WorkspaceCount: fmt.Sprint(len(target.Workspaces)),
			Default:        target.Default,
			Status:         views.GetStateLabel(target.State.Name),
		}

		if target.Metadata != nil {
			views_util.CheckAndAppendTimeLabel(&rowData.Status, target.State, target.Metadata.Uptime)
		}

		return getRowFromRowData(rowData)
	})

	footer := lipgloss.NewStyle().Foreground(views.LightGray).Render(views.GetListFooter(activeProfileName, &views.Padding{}))

	table := views_util.GetTableView(data, headers, &footer, func() {
		renderUnstyledList(targetList)
	})

	fmt.Println(table)
}

func SortTargets(targetList *[]apiclient.TargetDTO) {
	sort.Slice(*targetList, func(i, j int) bool {
		// Sort the default target on top
		if (*targetList)[i].Default && !(*targetList)[j].Default {
			return true
		}
		if !(*targetList)[i].Default && (*targetList)[j].Default {
			return false
		}

		pi, pj := views_util.GetStateSortPriorities((*targetList)[i].State.Name, (*targetList)[j].State.Name)
		if pi != pj {
			return pi < pj
		}

		// If two targets have the same state priority, compare the UpdatedAt property
		return (*targetList)[i].State.UpdatedAt > (*targetList)[j].State.UpdatedAt
	})
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
		isDefault = "Yes"
	} else {
		isDefault = "/"
	}

	return []string{
		fmt.Sprintf("%s %s", views.NameStyle.Render(rowData.Name), views.DefaultRowDataStyle.Render(fmt.Sprintf("(%s)", rowData.Provider))),
		views.DefaultRowDataStyle.Render(rowData.Options),
		views.DefaultRowDataStyle.Render(rowData.WorkspaceCount),
		isDefault,
		rowData.Status,
	}
}
