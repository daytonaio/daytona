// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/build/info"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"golang.org/x/term"
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

	re := lipgloss.NewRenderer(os.Stdout)

	headers := []string{"ID", "State", "Prebuild ID", "Created", "Updated"}

	data := [][]string{}

	for _, pc := range buildList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(pc)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(data)
		return
	}

	breakpointWidth := views.GetContainerBreakpointWidth(terminalWidth)

	minWidth := views_util.GetTableMinimumWidth(data)

	if breakpointWidth == 0 || minWidth > breakpointWidth {
		renderUnstyledList(buildList, apiServerConfig)
		return
	}

	t := table.New().
		Headers(headers...).
		Rows(data...).
		BorderStyle(re.NewStyle().Foreground(views.LightGray)).
		BorderRow(false).BorderColumn(false).BorderLeft(false).BorderRight(false).BorderTop(false).BorderBottom(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return views.TableHeaderStyle
			}
			return views.BaseCellStyle
		}).Width(breakpointWidth - 2*views.BaseTableStyleHorizontalPadding - 1)

	fmt.Println(views.BaseTableStyle.Render(t.String()))
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
