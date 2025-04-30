// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/daytonaio/daytona-ai-saas/cli/views/common"
	"github.com/daytonaio/daytona-ai-saas/cli/views/util"
	"github.com/daytonaio/daytona-ai-saas/daytonaapiclient"
)

type RowData struct {
	Name      string
	State     string
	Region    string
	Class     string
	LastEvent string
}

func ListSandboxes(sandboxList []daytonaapiclient.Workspace, activeOrganizationName *string) {
	if len(sandboxList) == 0 {
		util.NotifyEmptySandboxList(true)
		return
	}

	headers := []string{"Sandbox", "State", "Region", "Class", "Last Event"}

	data := [][]string{}

	for _, w := range sandboxList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(w)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table := util.GetTableView(data, headers, activeOrganizationName, func() {
		renderUnstyledList(sandboxList)
	})

	fmt.Println(table)
}

func SortSandboxes(sandboxList *[]daytonaapiclient.Workspace) {
	sort.Slice(*sandboxList, func(i, j int) bool {
		pi, pj := getStateSortPriorities(*(*sandboxList)[i].State, *(*sandboxList)[j].State)
		if pi != pj {
			return pi < pj
		}

		if (*sandboxList)[i].Info == nil || (*sandboxList)[j].Info == nil {
			return true
		}

		// If two sandboxes have the same state priority, compare the UpdatedAt property
		return (*sandboxList)[i].Info.Created > (*sandboxList)[j].Info.Created
	})
}

func getTableRowData(sandbox daytonaapiclient.Workspace) *RowData {
	rowData := RowData{"", "", "", "", ""}
	rowData.Name = sandbox.Id + util.AdditionalPropertyPadding
	if sandbox.State != nil {
		rowData.State = getStateLabel(*sandbox.State)
	}

	providerMetadataString := sandbox.Info.GetProviderMetadata()

	var providerMetadata providerMetadata

	err := json.Unmarshal([]byte(providerMetadataString), &providerMetadata)
	if err == nil {
		rowData.Region = providerMetadata.Region
		rowData.Class = providerMetadata.Class
		rowData.LastEvent = util.GetTimeSinceLabelFromString(providerMetadata.UpdatedAt)
	}

	return &rowData
}

func renderUnstyledList(sandboxList []daytonaapiclient.Workspace) {
	for _, sandbox := range sandboxList {
		RenderInfo(&sandbox, true)

		if sandbox.Id != sandboxList[len(sandboxList)-1].Id {
			fmt.Printf("\n%s\n\n", common.SeparatorString)
		}

	}
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		common.NameStyle.Render(rowData.Name),
		rowData.State,
		common.DefaultRowDataStyle.Render(rowData.Region),
		common.DefaultRowDataStyle.Render(rowData.Class),
		common.DefaultRowDataStyle.Render(rowData.LastEvent),
	}

	return row
}

func getStateSortPriorities(state1, state2 daytonaapiclient.WorkspaceState) (int, int) {
	pi, ok := sandboxListStatePriorities[state1]
	if !ok {
		pi = 99
	}
	pj, ok2 := sandboxListStatePriorities[state2]
	if !ok2 {
		pj = 99
	}

	return pi, pj
}

// Sandboxes that have actions being performed on them have a higher priority when listing
var sandboxListStatePriorities = map[daytonaapiclient.WorkspaceState]int{
	"pending":       1,
	"pending-start": 1,
	"deleting":      1,
	"creating":      1,
	"started":       2,
	"undefined":     2,
	"error":         3,
	"stopped":       4,
}
