// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"os"
	"sort"

	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

type RowData struct {
	Target   string
	Provider string
	Options  string
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.Target),
		views.DefaultRowDataStyle.Render(rowData.Provider),
		views.DefaultRowDataStyle.Render(rowData.Options),
	}

	return row
}

func getRowData(target *serverapiclient.ProviderTarget) *RowData {
	rowData := RowData{"", "", ""}

	rowData.Target = *target.Name
	rowData.Provider = *target.ProviderInfo.Name
	rowData.Options = *target.Options

	return &rowData
}

func ListTargets(targetList []serverapiclient.ProviderTarget) {

	sortTargets(&targetList)

	re := lipgloss.NewRenderer(os.Stdout)

	headers := []string{"Target", "Provider", "Options"}

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

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(data)
		return
	}

	breakpointWidth := views.GetContainerBreakpointWidth(terminalWidth)

	if breakpointWidth == 0 || terminalWidth < views.TUITableMinimumWidth {
		renderUnstyledList(targetList)
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
		}).Width(breakpointWidth - 2*views.BaseTableStyleHorizontalPadding)

	fmt.Println(views.BaseTableStyle.Render(t.String()))
}

func sortTargets(targets *[]serverapiclient.ProviderTarget) {
	sort.Slice(*targets, func(i, j int) bool {
		t1 := (*targets)[i]
		t2 := (*targets)[j]
		return *t1.ProviderInfo.Name < *t2.ProviderInfo.Name
	})
}

func renderUnstyledList(targetList []serverapiclient.ProviderTarget) {
	output := "\n"

	for _, target := range targetList {
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Target Name: "), *target.Name) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Target Provider: "), *target.ProviderInfo.Name) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Target Options: "), *target.Options) + "\n"

		if target.Name != targetList[len(targetList)-1].Name {
			output += views.SeparatorString + "\n\n"
		}
	}

	fmt.Println(output)
}
