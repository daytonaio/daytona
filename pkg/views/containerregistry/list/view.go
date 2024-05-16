// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

type rowData struct {
	Server   string
	Username string
	Password string
}

func getRowData(registry *apiclient.ContainerRegistry) *rowData {
	rowData := rowData{"", "", ""}

	rowData.Server = *registry.Server
	rowData.Username = *registry.Username
	rowData.Password = *registry.Password

	return &rowData
}

func getRowFromRowData(rowData rowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.Server),
		views.DefaultRowDataStyle.Render(rowData.Username),
		views.DefaultRowDataStyle.Render(rowData.Password),
	}

	return row
}

func ListRegistries(registryList []apiclient.ContainerRegistry) {
	re := lipgloss.NewRenderer(os.Stdout)
	headers := []string{"Server", "Username", "Password"}
	data := [][]string{}

	for _, registry := range registryList {
		var rowData *rowData
		var row []string

		rowData = getRowData(&registry)
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
	minWidth := views_util.GetTableMinimumWidth(data)
	if breakpointWidth == 0 || terminalWidth < views.TUITableMinimumWidth || minWidth > breakpointWidth {
		renderUnstyledList(registryList)
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

func renderUnstyledList(registryList []apiclient.ContainerRegistry) {
	output := "\n"

	for _, registry := range registryList {
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Server: "), *registry.Server) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Username: "), *registry.Username) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Password: "), *registry.Password) + "\n\n"

		if registry.Server != registryList[len(registryList)-1].Server {
			output += views.SeparatorString + "\n\n"
		}
	}

	fmt.Println(output)
}
