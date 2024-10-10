// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

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
	Label   string
	Name    string
	Version string
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.Label),
		views.DefaultRowDataStyle.Render(rowData.Name),
		views.DefaultRowDataStyle.Render(rowData.Version),
	}

	return row
}

func getRowData(provider *apiclient.Provider) *RowData {
	rowData := RowData{"", "", ""}

	if provider.Label != nil {
		rowData.Label = *provider.Label
	} else {
		rowData.Label = provider.Name
	}
	rowData.Name = provider.Name
	rowData.Version = provider.Version

	return &rowData
}

func List(providerList []apiclient.Provider) {

	re := lipgloss.NewRenderer(os.Stdout)

	headers := []string{"Provider", "Name", "Version"}

	data := [][]string{}

	for _, provider := range providerList {
		var rowData *RowData
		var row []string

		rowData = getRowData(&provider)
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
		renderUnstyledList(providerList)
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

func renderUnstyledList(providerList []apiclient.Provider) {
	output := "\n"

	for _, provider := range providerList {
		if provider.Label != nil {
			output += fmt.Sprintf("%s %s", views.GetPropertyKey("Provider: "), *provider.Label) + "\n\n"
		}
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Name: "), provider.Name) + "\n\n"
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Version: "), provider.Version) + "\n"

		if provider.Name != providerList[len(providerList)-1].Name {
			output += views.SeparatorString + "\n\n"
		}
	}

	fmt.Println(output)
}
