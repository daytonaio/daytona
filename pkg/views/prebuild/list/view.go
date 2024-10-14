// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"strconv"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/prebuild/info"
	"github.com/daytonaio/daytona/pkg/views/util"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

var maxTriggerFilesStringLength = 24

type RowData struct {
	ProjectConfigName string
	Branch            string
	CommitInterval    string
	TriggerFiles      string
	Retention         string
}

func ListPrebuilds(prebuildList []apiclient.PrebuildDTO) {
	data := [][]string{}

	for _, pc := range prebuildList {
		var rowData *RowData
		var row []string

		rowData = getTableRowData(pc)
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table, success := util.GetTableView(data, []string{
		"Project Config", "Branch", "Commit Interval", "Trigger files", "Build Retention",
	}, nil)

	if !success {
		renderUnstyledList(prebuildList)
		return
	}

	fmt.Println(table)
}

func renderUnstyledList(prebuildList []apiclient.PrebuildDTO) {
	for _, pc := range prebuildList {
		info.Render(&pc, true)

		if pc.Id != prebuildList[len(prebuildList)-1].Id {
			fmt.Printf("\n%s\n\n", views.SeparatorString)
		}
	}
}

func getRowFromRowData(rowData RowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.ProjectConfigName),
		views.DefaultRowDataStyle.Render(views.GetBranchNameLabel(rowData.Branch)),
		views.ActiveStyle.Render(rowData.CommitInterval),
		views.DefaultRowDataStyle.Render(rowData.TriggerFiles),
		views.DefaultRowDataStyle.Render(rowData.Retention),
	}

	return row
}

func getTableRowData(prebuildConfig apiclient.PrebuildDTO) *RowData {
	rowData := RowData{"", "", "", "", ""}

	rowData.ProjectConfigName = prebuildConfig.ProjectConfigName + views_util.AdditionalPropertyPadding
	rowData.Branch = prebuildConfig.Branch
	if prebuildConfig.CommitInterval != nil {
		rowData.CommitInterval = strconv.Itoa(int(*prebuildConfig.CommitInterval))
	} else {
		rowData.CommitInterval = views.InactiveStyle.Render("None")
	}
	rowData.TriggerFiles = getTriggerFilesString(prebuildConfig.TriggerFiles)
	rowData.Retention = strconv.Itoa(int(prebuildConfig.Retention))

	return &rowData
}

func getTriggerFilesString(triggerFiles []string) string {
	if len(triggerFiles) == 0 {
		return views.InactiveStyle.Render("None")
	}

	var fileString string
	result := "[ "

	for i, triggerFile := range triggerFiles {
		fileString += triggerFile
		if i != len(triggerFiles)-1 {
			fileString += ", "
		}
	}

	if len(fileString) > maxTriggerFilesStringLength {
		fileString = fileString[:maxTriggerFilesStringLength-3] + "..."
	}

	result += fileString
	result += " ]"

	return result
}
