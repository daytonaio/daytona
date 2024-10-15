// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type rowData struct {
	ID     string
	Name   string
	Status string
	ApiUrl string
	ApiKey string
}

func ListProfiles(profileList []config.Profile, activeProfileId string, showApiKeysFlag bool) (string, error) {
	headers := []string{"ID", "Name", "Status", "API URL"}
	if showApiKeysFlag {
		headers = append(headers, "API Key")
	}

	data := [][]string{}

	for _, profile := range profileList {
		data = append(data, getRowFromData(&profile, activeProfileId, showApiKeysFlag))
	}

	table, success := views_util.GetTableView(data, headers, nil)

	if !success {
		return renderUnstyledList(profileList, activeProfileId, showApiKeysFlag), nil
	}

	return table + "\n", nil
}

func getRowFromData(profile *config.Profile, activeProfileId string, showApiKeysFlag bool) []string {
	var data rowData

	data.ID = profile.Id
	data.Name = profile.Name
	data.ApiUrl = profile.Api.Url
	if profile.Id == activeProfileId {
		data.Status = "1"
	}
	if showApiKeysFlag {
		data.ApiKey = profile.Api.Key
	}

	var state string
	if data.Status == "" {
		state = views.InactiveStyle.Render("Inactive")
	} else {
		state = views.ActiveStyle.Render("Active")
	}

	row := []string{
		views.NameStyle.Render(data.ID),
		views.DefaultRowDataStyle.Render(data.Name),
		state,
		views.DefaultRowDataStyle.Render(data.ApiUrl),
	}
	if showApiKeysFlag {
		row = append(row, views.DefaultRowDataStyle.Render(data.ApiKey))
	}

	return row
}

func renderUnstyledList(profileList []config.Profile, activeProfileId string, showApiKeysFlag bool) string {
	var status string
	var isActive bool

	output := "\n"

	for _, profile := range profileList {
		if profile.Id == activeProfileId {
			isActive = true
		}
		if isActive {
			status = views.ActiveStyle.Render("Active")
		} else {
			status = views.InactiveStyle.Render("Inactive")
		}

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Profile Name: "), profile.Name) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Profile ID: "), profile.Id) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Status: "), status) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("API URL: "), profile.Api.Url) + "\n\n"

		if showApiKeysFlag {
			output += fmt.Sprintf("%s %s", views.GetPropertyKey("API Key: "), profile.Api.Key) + "\n\n"
		}

		if profile.Id != profileList[len(profileList)-1].Id {
			output += views.SeparatorString + "\n\n"
		}

		isActive = false
	}

	return output
}
