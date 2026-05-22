// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"fmt"
	"sort"

	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/util"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

type RowData struct {
	Name      string
	State     string
	Region    string
	LastEvent string
}

func ListSandboxes(sandboxList []apiclient.SandboxListItem, activeOrganizationName *string) {
	if len(sandboxList) == 0 {
		util.NotifyEmptySandboxList(true)
		return
	}

	headers := []string{"Sandbox", "State", "Region", "Last Event"}

	data := [][]string{}

	for _, s := range sandboxList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(s)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table := util.GetTableView(data, headers, activeOrganizationName, func() {
		renderUnstyledList(sandboxList)
	})

	fmt.Println(table)
}

func SortSandboxes(sandboxList *[]apiclient.SandboxListItem) {
	sort.Slice(*sandboxList, func(i, j int) bool {
		pi, pj := getStateSortPriorities(*(*sandboxList)[i].State, *(*sandboxList)[j].State)
		if pi != pj {
			return pi < pj
		}

		if (*sandboxList)[i].CreatedAt == nil || (*sandboxList)[j].CreatedAt == nil {
			return true
		}

		// If two sandboxes have the same state priority, compare the CreatedAt property
		return *(*sandboxList)[i].CreatedAt > *(*sandboxList)[j].CreatedAt
	})
}

func getTableRowData(sandbox apiclient.SandboxListItem) *RowData {
	rowData := RowData{"", "", "", ""}
	rowData.Name = sandbox.Id + util.AdditionalPropertyPadding
	if sandbox.State != nil {
		rowData.State = getStateLabel(*sandbox.State)
	}

	rowData.Region = sandbox.Target

	if sandbox.LastActivityAt != nil {
		rowData.LastEvent = util.GetTimeSinceLabelFromString(*sandbox.LastActivityAt)
	}

	return &rowData
}

func renderUnstyledList(sandboxList []apiclient.SandboxListItem) {
	for i := range sandboxList {
		// Pass by pointer so the value implements the sandboxInfo interface
		// (pointer receiver methods on *apiclient.SandboxListItem).
		RenderInfo(&sandboxList[i], true)

		if sandboxList[i].Id != sandboxList[len(sandboxList)-1].Id {
			fmt.Printf("\n%s\n\n", common.SeparatorString)
		}
	}
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		common.NameStyle.Render(rowData.Name),
		rowData.State,
		common.DefaultRowDataStyle.Render(rowData.Region),
		common.DefaultRowDataStyle.Render(rowData.LastEvent),
	}

	return row
}

func getStateSortPriorities(state1, state2 apiclient.SandboxState) (int, int) {
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
var sandboxListStatePriorities = map[apiclient.SandboxState]int{
	"pending":       1,
	"pending-start": 1,
	"deleting":      1,
	"creating":      1,
	"started":       2,
	"undefined":     2,
	"error":         3,
	"build-failed":  3,
	"stopped":       4,
}
