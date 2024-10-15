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

type rowData struct {
	ProjectConfigName string
	Branch            string
	CommitInterval    string
	TriggerFiles      string
	Retention         string
}

func ListPrebuilds(prebuildList []apiclient.PrebuildDTO) {
	data := [][]string{}

	for _, p := range prebuildList {
		data = append(data, getRowFromData(p))
	}

	table := util.GetTableView(data, []string{
		"Project Config", "Branch", "Commit Interval", "Trigger files", "Build Retention",
	}, nil, renderUnstyledList, prebuildList)

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

func getRowFromData(prebuildConfig apiclient.PrebuildDTO) []string {
	var data rowData

	data.ProjectConfigName = prebuildConfig.ProjectConfigName + views_util.AdditionalPropertyPadding
	data.Branch = prebuildConfig.Branch
	if prebuildConfig.CommitInterval != nil {
		data.CommitInterval = strconv.Itoa(int(*prebuildConfig.CommitInterval))
	} else {
		data.CommitInterval = views.InactiveStyle.Render("None")
	}
	data.TriggerFiles = getTriggerFilesString(prebuildConfig.TriggerFiles)
	data.Retention = strconv.Itoa(int(prebuildConfig.Retention))

	return []string{
		views.NameStyle.Render(data.ProjectConfigName),
		views.DefaultRowDataStyle.Render(views.GetBranchNameLabel(data.Branch)),
		views.ActiveStyle.Render(data.CommitInterval),
		views.DefaultRowDataStyle.Render(data.TriggerFiles),
		views.DefaultRowDataStyle.Render(data.Retention),
	}
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
