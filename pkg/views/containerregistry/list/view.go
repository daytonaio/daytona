// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type rowData struct {
	Server   string
	Username string
	Password string
}

func ListRegistries(registryList []apiclient.ContainerRegistry) {
	data := [][]string{}

	for _, registry := range registryList {
		var rowData *rowData
		var row []string

		rowData = getRowData(&registry)
		if rowData == nil {
			continue
		}
		row = getRowFromRowData(*rowData)
		data = append(data, row)
	}

	table, success := views_util.GetTableView(data, []string{
		"Server", "Username", "Password",
	}, nil)

	if !success {
		renderUnstyledList(registryList)
		return
	}

	fmt.Println(table)
}

func getRowData(registry *apiclient.ContainerRegistry) *rowData {
	rowData := rowData{"", "", ""}

	rowData.Server = registry.Server
	rowData.Username = registry.Username
	rowData.Password = registry.Password

	return &rowData
}

func getRowFromRowData(rowData rowData) []string {
	row := []string{
		views.NameStyle.Render(rowData.Server),
		views.DefaultRowDataStyle.Render(rowData.Username),
		views.DefaultRowDataStyle.Render(strings.Repeat("*", 10)),
	}

	return row
}

func renderUnstyledList(registryList []apiclient.ContainerRegistry) {
	output := "\n"

	for _, registry := range registryList {
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Server: "), registry.Server) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Username: "), registry.Username) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Password: "), registry.Password) + "\n\n"

		if registry.Server != registryList[len(registryList)-1].Server {
			output += views.SeparatorString + "\n\n"
		}
	}

	fmt.Println(output)
}
