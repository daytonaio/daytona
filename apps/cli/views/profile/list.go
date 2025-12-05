// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/util"
)

type RowData struct {
	Name   string
	Id     string
	ApiUrl string
}

func ListProfiles(profileList []config.Profile, activeProfileId *string) {
	if len(profileList) == 0 {
		util.NotifyEmptyProfileList(true)
		return
	}

	headers := []string{"Name", "ID", "API URL"}

	data := [][]string{}

	for _, p := range profileList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(p, activeProfileId)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table := util.GetTableView(data, headers, nil, func() {
		renderUnstyledList(profileList, activeProfileId)
	})

	fmt.Println(table)
}

func getTableRowData(profile config.Profile, activeProfileId *string) *RowData {
	rowData := RowData{"", "", ""}
	rowData.Name = profile.Name + util.AdditionalPropertyPadding

	if activeProfileId != nil && *activeProfileId == profile.Id {
		rowData.Name = "*" + rowData.Name
	}

	rowData.Id = profile.Id
	rowData.ApiUrl = profile.Api.Url

	return &rowData
}

func renderUnstyledList(profileList []config.Profile, activeProfileId *string) {
	for _, profile := range profileList {
		RenderInfo(&profile, activeProfileId, true)

		if profile.Id != profileList[len(profileList)-1].Id {
			fmt.Printf("\n%s\n\n", common.SeparatorString)
		}
	}
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		common.NameStyle.Render(rowData.Name),
		common.DefaultRowDataStyle.Render(rowData.Id),
		common.DefaultRowDataStyle.Render(rowData.ApiUrl),
	}

	return row
}
