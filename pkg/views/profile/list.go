// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

type RowData struct {
	ID     string
	Name   string
	Status string
	ApiUrl string
}

func getRowFromRowData(rowData RowData) []string {
	var state string
	if rowData.Status == "" {
		state = views.InactiveStyle.Render("Inactive")
	} else {
		state = views.ActiveStyle.Render("Active")
	}

	row := []string{
		views.NameStyle.Render(rowData.ID),
		views.DefaultRowDataStyle.Render(rowData.Name),
		state,
		views.DefaultRowDataStyle.Render(rowData.ApiUrl),
	}

	return row
}

func getRowData(profile *config.Profile, activeProfileId string) *RowData {
	rowData := RowData{"", "", "", ""}

	rowData.ID = profile.Id
	rowData.Name = profile.Name
	rowData.ApiUrl = profile.Api.Url
	if profile.Id == activeProfileId {
		rowData.Status = "1"
	}

	return &rowData
}

func ListProfiles(profileList []config.Profile, activeProfileId string) {

	re := lipgloss.NewRenderer(os.Stdout)

	headers := []string{"ID", "Name", "Status", "API URL"}

	data := [][]string{}

	for _, profile := range profileList {
		var rowData *RowData
		var row []string

		rowData = getRowData(&profile, activeProfileId)
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
		renderUnstyledList(profileList, activeProfileId)
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

func renderUnstyledList(profileList []config.Profile, activeProfileId string) {
	var status string
	var isActive bool

	output := "\n"

	for _, profile := range profileList {
		if profile.Id == activeProfileId {
			isActive = true
		}
		if isActive {
			status = views.ActiveStyle.Render("Active")
		} else {
			status = views.InactiveStyle.Render("Inactive")
		}

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Profile Name: "), profile.Name) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Profile ID: "), profile.Id) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Status: "), status) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("API URL: "), profile.Api.Url) + "\n\n"

		if profile.Id != profileList[len(profileList)-1].Id {
			output += views.SeparatorString + "\n\n"
		}

		isActive = false
	}

	fmt.Println(output)
}
