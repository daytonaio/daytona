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
		data = append(data, getRowFromData(registry))
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

func getRowFromData(registry apiclient.ContainerRegistry) []string {
	var data rowData

	data.Server = registry.Server
	data.Username = registry.Username

	row := []string{
		views.NameStyle.Render(data.Server),
		views.DefaultRowDataStyle.Render(data.Username),
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
