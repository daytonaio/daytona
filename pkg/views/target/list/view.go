// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"sort"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"
)

type rowData struct {
	Target    string
	Provider  string
	IsDefault string
	Options   string
}

func ListTargets(targetList []apiclient.ProviderTarget) {
	sortTargets(&targetList)

	data := [][]string{}

	for _, target := range targetList {
		data = append(data, getRowFromRowData(&target))
	}

	table := util.GetTableView(data, []string{
		"Target", "Provider", "Default", "Options",
	}, nil, func() {
		renderUnstyledList(targetList)
	})

	fmt.Println(table)
}

func getRowFromRowData(target *apiclient.ProviderTarget) []string {
	var isDefault string
	var data rowData

	data.Target = target.Name
	data.Provider = target.ProviderInfo.Name
	data.Options = target.Options

	if target.IsDefault {
		isDefault = views.ActiveStyle.Render("Yes")
	} else {
		isDefault = views.InactiveStyle.Render("/")
	}

	row := []string{
		views.NameStyle.Render(data.Target),
		views.DefaultRowDataStyle.Render(data.Provider),
		isDefault,
		views.DefaultRowDataStyle.Render(data.Options),
	}

	return row
}

func sortTargets(targets *[]apiclient.ProviderTarget) {
	sort.Slice(*targets, func(i, j int) bool {
		t1 := (*targets)[i]
		t2 := (*targets)[j]
		return t1.ProviderInfo.Name < t2.ProviderInfo.Name
	})
}

func renderUnstyledList(targetList []apiclient.ProviderTarget) {
	output := "\n"

	for _, target := range targetList {
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Target Name: "), target.Name) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Target Provider: "), target.ProviderInfo.Name) + "\n\n"

		if target.IsDefault {
			output += fmt.Sprintf("%s %s", views.GetPropertyKey("Default: "), "Yes") + "\n\n"
		}

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Target Options: "), target.Options) + "\n\n"

		if target.Name != targetList[len(targetList)-1].Name {
			output += views.SeparatorString + "\n\n"
		}
	}

	fmt.Println(output)
}
