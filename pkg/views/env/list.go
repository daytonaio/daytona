// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"golang.org/x/term"
)

type RowData struct {
	Key   string
	Value string
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.Key),
		views.DefaultRowDataStyle.Render(rowData.Value),
	}

	return row
}

func getRowData(key, value string) *RowData {
	return &RowData{key, value}
}

func List(envVars []apiclient.EnvironmentVariable) {
	if len(envVars) == 0 {
		views_util.NotifyEmptyEnvVarList(true)
		return
	}

	re := lipgloss.NewRenderer(os.Stdout)

	headers := []string{"Key", "Value"}

	data := [][]string{}

	for _, envVar := range envVars {
		var rowData *RowData
		var row []string

		rowData = getRowData(envVar.Key, envVar.Value)
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
		renderUnstyledList(envVars)
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

func renderUnstyledList(envVars []apiclient.EnvironmentVariable) {
	output := "\n"

	for _, envVar := range envVars {
		output += fmt.Sprintf("%s\t%s", views.GetPropertyKey("Key:"), envVar.Key) + "\n"
		output += fmt.Sprintf("%s\t%s", views.GetPropertyKey("Value:"), envVar.Value) + "\n"

		output += "\n\n"
	}

	output = strings.TrimSuffix(output, "\n\n")

	fmt.Println(output)
}
