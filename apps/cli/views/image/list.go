// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package image

import (
	"fmt"
	"sort"

	"github.com/daytonaio/daytona-ai-saas/cli/views/common"
	"github.com/daytonaio/daytona-ai-saas/cli/views/util"
	"github.com/daytonaio/daytona-ai-saas/daytonaapiclient"
)

type RowData struct {
	Name    string
	State   string
	Enabled string
	Size    string
	Created string
}

func ListImages(imageList []daytonaapiclient.ImageDto, activeOrganizationName *string) {
	if len(imageList) == 0 {
		util.NotifyEmptyImageList(true)
		return
	}

	SortImages(&imageList)

	headers := []string{"Image", "State", "Enabled", "Size", "Created"}

	data := [][]string{}

	for _, img := range imageList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(img)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table := util.GetTableView(data, headers, activeOrganizationName, func() {
		renderUnstyledList(imageList)
	})

	fmt.Println(table)
}

func SortImages(imageList *[]daytonaapiclient.ImageDto) {
	sort.Slice(*imageList, func(i, j int) bool {
		pi, pj := getStateSortPriorities((*imageList)[i].State, (*imageList)[j].State)
		if pi != pj {
			return pi < pj
		}

		// If two images have the same state priority, compare the UpdatedAt property
		return (*imageList)[i].CreatedAt.After((*imageList)[j].CreatedAt)
	})
}

func getTableRowData(image daytonaapiclient.ImageDto) *RowData {
	rowData := RowData{"", "", "", "", ""}
	rowData.Name = image.Name + util.AdditionalPropertyPadding
	rowData.State = getStateLabel(image.State)

	if image.Enabled {
		rowData.Enabled = "Yes"
	} else {
		rowData.Enabled = "No"
	}

	if image.Size.IsSet() && image.Size.Get() != nil {
		rowData.Size = fmt.Sprintf("%.2f GB", *image.Size.Get())
	} else {
		rowData.Size = "-"
	}

	rowData.Created = util.GetTimeSinceLabel(image.CreatedAt)
	return &rowData
}

func renderUnstyledList(imageList []daytonaapiclient.ImageDto) {
	for _, image := range imageList {
		RenderInfo(&image, true)

		if image.Id != imageList[len(imageList)-1].Id {
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

func getStateSortPriorities(state1, state2 daytonaapiclient.ImageState) (int, int) {
	pi, ok := imageListStatePriorities[state1]
	if !ok {
		pi = 99
	}
	pj, ok2 := imageListStatePriorities[state2]
	if !ok2 {
		pj = 99
	}

	return pi, pj
}

// Images that have actions being performed on them have a higher priority when listing
var imageListStatePriorities = map[daytonaapiclient.ImageState]int{
	daytonaapiclient.IMAGESTATE_PENDING:       1,
	daytonaapiclient.IMAGESTATE_PULLING_IMAGE: 1,
	daytonaapiclient.IMAGESTATE_VALIDATING:    1,
	daytonaapiclient.IMAGESTATE_ERROR:         2,
	daytonaapiclient.IMAGESTATE_ACTIVE:        3,
	daytonaapiclient.IMAGESTATE_REMOVING:      4,
}
