// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"fmt"
	"sort"

	"github.com/daytonaio/apiclient"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/util"
)

type RowData struct {
	Name    string
	State   string
	Enabled string
	Size    string
	Created string
}

func ListSnapshots(snapshotList []apiclient.SnapshotDto, activeOrganizationName *string) {
	if len(snapshotList) == 0 {
		util.NotifyEmptySnapshotList(true)
		return
	}

	SortSnapshots(&snapshotList)

	headers := []string{"Snapshot", "State", "Enabled", "Size", "Created"}

	data := [][]string{}

	for _, img := range snapshotList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(img)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table := util.GetTableView(data, headers, activeOrganizationName, func() {
		renderUnstyledList(snapshotList)
	})

	fmt.Println(table)
}

func SortSnapshots(snapshotList *[]apiclient.SnapshotDto) {
	sort.Slice(*snapshotList, func(i, j int) bool {
		pi, pj := getStateSortPriorities((*snapshotList)[i].State, (*snapshotList)[j].State)
		if pi != pj {
			return pi < pj
		}

		// If two snapshots have the same state priority, compare the UpdatedAt property
		return (*snapshotList)[i].CreatedAt.After((*snapshotList)[j].CreatedAt)
	})
}

func getTableRowData(snapshot apiclient.SnapshotDto) *RowData {
	rowData := RowData{"", "", "", "", ""}
	rowData.Name = snapshot.Name + util.AdditionalPropertyPadding
	rowData.State = getStateLabel(snapshot.State)

	if snapshot.Enabled {
		rowData.Enabled = "Yes"
	} else {
		rowData.Enabled = "No"
	}

	if snapshot.Size.IsSet() && snapshot.Size.Get() != nil {
		rowData.Size = fmt.Sprintf("%.2f GB", *snapshot.Size.Get())
	} else {
		rowData.Size = "-"
	}

	rowData.Created = util.GetTimeSinceLabel(snapshot.CreatedAt)
	return &rowData
}

func renderUnstyledList(snapshotList []apiclient.SnapshotDto) {
	for _, snapshot := range snapshotList {
		RenderInfo(&snapshot, true)

		if snapshot.Id != snapshotList[len(snapshotList)-1].Id {
			fmt.Printf("\n%s\n\n", common.SeparatorString)
		}
	}
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		common.NameStyle.Render(rowData.Name),
		rowData.State,
		common.DefaultRowDataStyle.Render(rowData.Enabled),
		common.DefaultRowDataStyle.Render(rowData.Size),
		common.DefaultRowDataStyle.Render(rowData.Created),
	}

	return row
}

func getStateSortPriorities(state1, state2 apiclient.SnapshotState) (int, int) {
	pi, ok := snapshotListStatePriorities[state1]
	if !ok {
		pi = 99
	}
	pj, ok2 := snapshotListStatePriorities[state2]
	if !ok2 {
		pj = 99
	}

	return pi, pj
}

// snapshots that have actions being performed on them have a higher priority when listing
var snapshotListStatePriorities = map[apiclient.SnapshotState]int{
	apiclient.SNAPSHOTSTATE_PENDING:    1,
	apiclient.SNAPSHOTSTATE_PULLING:    1,
	apiclient.SNAPSHOTSTATE_VALIDATING: 1,
	apiclient.SNAPSHOTSTATE_ERROR:      2,
	apiclient.SNAPSHOTSTATE_ACTIVE:     3,
	apiclient.SNAPSHOTSTATE_REMOVING:   4,
}
