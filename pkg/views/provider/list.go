// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"
)

type rowData struct {
	Label   string
	Name    string
	Version string
}

func List(providerList []apiclient.Provider) {
	data := [][]string{}

	for _, provider := range providerList {
		data = append(data, getRowFromData(&provider))
	}

	table := util.GetTableView(data, []string{
		"Provider", "Name", "Version",
	}, nil, func() {
		renderUnstyledList(providerList)
	})

	fmt.Println(table)
}

func getRowFromData(provider *apiclient.Provider) []string {
	var data rowData

	if provider.Label != nil {
		data.Label = *provider.Label
	} else {
		data.Label = provider.Name
	}
	data.Name = provider.Name
	data.Version = provider.Version

	return []string{
		views.NameStyle.Render(data.Label),
		views.DefaultRowDataStyle.Render(data.Name),
		views.DefaultRowDataStyle.Render(data.Version),
	}
}

func renderUnstyledList(providerList []apiclient.Provider) {
	output := "\n"

	for _, provider := range providerList {
		if provider.Label != nil {
			output += fmt.Sprintf("%s %s", views.GetPropertyKey("Provider: "), *provider.Label) + "\n\n"
		}
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Name: "), provider.Name) + "\n\n"
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Version: "), provider.Version) + "\n"

		if provider.Name != providerList[len(providerList)-1].Name {
			output += views.SeparatorString + "\n\n"
		}
	}

	fmt.Println(output)
}
