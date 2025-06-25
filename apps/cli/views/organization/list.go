// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package organization

import (
	"fmt"
	"sort"

	"github.com/daytonaio/apiclient"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/util"
)

type RowData struct {
	Name    string
	Id      string
	Created string
}

func ListOrganizations(organizationList []apiclient.Organization, activeOrganizationId *string) {
	if len(organizationList) == 0 {
		util.NotifyEmptyOrganizationList(true)
		return
	}

	SortOrganizations(&organizationList)

	headers := []string{"Name", "Id", "Created"}

	data := [][]string{}

	for _, o := range organizationList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(o, activeOrganizationId)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table := util.GetTableView(data, headers, nil, func() {
		renderUnstyledList(organizationList)
	})

	fmt.Println(table)
}

func SortOrganizations(organizationList *[]apiclient.Organization) {
	sort.Slice(*organizationList, func(i, j int) bool {
		return (*organizationList)[i].CreatedAt.After((*organizationList)[j].CreatedAt)
	})
}

func getTableRowData(organization apiclient.Organization, activeOrganizationId *string) *RowData {
	rowData := RowData{"", "", ""}
	rowData.Name = organization.Name + util.AdditionalPropertyPadding

	if activeOrganizationId != nil && *activeOrganizationId == organization.Id {
		rowData.Name = "*" + rowData.Name
	}

	rowData.Id = organization.Id
	rowData.Created = util.GetTimeSinceLabel(organization.CreatedAt)

	return &rowData
}

func renderUnstyledList(organizationList []apiclient.Organization) {
	for _, organization := range organizationList {
		RenderInfo(&organization, true)

		if organization.Id != organizationList[len(organizationList)-1].Id {
			fmt.Printf("\n%s\n\n", common.SeparatorString)
		}

	}
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		common.NameStyle.Render(rowData.Name),
		common.DefaultRowDataStyle.Render(rowData.Id),
		common.DefaultRowDataStyle.Render(rowData.Created),
	}

	return row
}
