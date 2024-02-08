// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile_info

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type ProfileInfo struct {
	ProfileName string
	ServerUrl   string
	ProfilePath string
}

var whiteText = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#eeeeee"))

var grayText = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#eeeeee"))

func Render(info ProfileInfo, verb string) {
	output := ""
	output = grayText.Render("Profile ") + whiteText.Render(info.ProfileName) + grayText.Render(fmt.Sprintf(" %s", verb)) + "\n"
	output += grayText.Render("Server URL: ") + whiteText.Render(info.ServerUrl)

	println(output)
}
