// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

type RowData struct {
	Name string
	Type string
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.Name),
		views.DefaultRowDataStyle.Render(rowData.Type),
	}

	return row
}

func getRowData(apiKey *apiclient.ApiKey) *RowData {
	rowData := RowData{"", ""}

	rowData.Name = apiKey.Name
	rowData.Type = string(apiKey.Type)

	return &rowData
}

func ListApiKeys(apiKeyList []apiclient.ApiKey) {

	re := lipgloss.NewRenderer(os.Stdout)

	headers := []string{"Name", "Type"}

	data := [][]string{}

	for _, apiKey := range apiKeyList {
		var rowData *RowData
		var row []string

		rowData = getRowData(&apiKey)
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
		renderUnstyledList(apiKeyList)
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

func renderUnstyledList(apiKeyList []apiclient.ApiKey) {
	output := "\n"

	for _, apiKey := range apiKeyList {
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("API Key Name: "), apiKey.Name) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("API Key Type: "), apiKey.Type) + "\n\n"

		if apiKey.Name != apiKeyList[len(apiKeyList)-1].Name {
			output += views.SeparatorString + "\n\n"
		}
	}

	fmt.Println(output)
}
