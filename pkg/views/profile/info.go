// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type ProfileInfo struct {
	ProfileName string
	ApiUrl      string
	ProfilePath string
}

var whiteText = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#eeeeee"))

var grayText = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#eeeeee"))

func Render(info ProfileInfo, verb string) {
	if info.ApiUrl == "" {
		info.ApiUrl = "N/A - set by editing the profile later"
	}

	output := ""
	output = grayText.Render("Profile ") + whiteText.Render(info.ProfileName) + grayText.Render(fmt.Sprintf(" %s", verb)) + "\n"
	output += grayText.Render("Server URL: ") + whiteText.Render(info.ApiUrl)

	println(output)
}
