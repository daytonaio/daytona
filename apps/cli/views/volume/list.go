// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

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
	Size    string
	Created string
}

func ListVolumes(volumeList []apiclient.VolumeDto, activeOrganizationName *string) {
	if len(volumeList) == 0 {
		util.NotifyEmptyVolumeList(true)
		return
	}

	SortVolumes(&volumeList)

	headers := []string{"Volume", "State", "Size", "Created"}

	data := [][]string{}

	for _, v := range volumeList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(v)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table := util.GetTableView(data, headers, activeOrganizationName, func() {
		renderUnstyledList(volumeList)
	})

	fmt.Println(table)
}

func SortVolumes(volumeList *[]apiclient.VolumeDto) {
	sort.Slice(*volumeList, func(i, j int) bool {
		if (*volumeList)[i].State != (*volumeList)[j].State {
			pi, pj := getStateSortPriorities((*volumeList)[i].State, (*volumeList)[j].State)
			return pi < pj
		}

		// If two volumes have the same state priority, compare the CreatedAt property
		return (*volumeList)[i].CreatedAt > (*volumeList)[j].CreatedAt
	})
}

func getTableRowData(volume apiclient.VolumeDto) *RowData {
	rowData := RowData{"", "", "", ""}
	rowData.Name = volume.Name + util.AdditionalPropertyPadding
	rowData.State = getStateLabel(volume.State)
	rowData.Created = util.GetTimeSinceLabelFromString(volume.CreatedAt)
	return &rowData
}

func renderUnstyledList(volumeList []apiclient.VolumeDto) {
	for _, volume := range volumeList {
		RenderInfo(&volume, true)

		if volume.Id != volumeList[len(volumeList)-1].Id {
			fmt.Printf("\n%s\n\n", common.SeparatorString)
		}
	}
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		common.NameStyle.Render(rowData.Name),
		rowData.State,
		common.DefaultRowDataStyle.Render(rowData.Size),
		common.DefaultRowDataStyle.Render(rowData.Created),
	}

	return row
}

func getStateSortPriorities(state1, state2 apiclient.VolumeState) (int, int) {
	pi, ok := volumeListStatePriorities[state1]
	if !ok {
		pi = 99
	}
	pj, ok2 := volumeListStatePriorities[state2]
	if !ok2 {
		pj = 99
	}

	return pi, pj
}

// Volumes that have actions being performed on them have a higher priority when listing
var volumeListStatePriorities = map[apiclient.VolumeState]int{
	apiclient.VOLUMESTATE_PENDING_CREATE: 1,
	apiclient.VOLUMESTATE_CREATING:       1,
	apiclient.VOLUMESTATE_PENDING_DELETE: 1,
	apiclient.VOLUMESTATE_DELETING:       1,
	apiclient.VOLUMESTATE_READY:          2,
	apiclient.VOLUMESTATE_ERROR:          3,
	apiclient.VOLUMESTATE_DELETED:        4,
}
