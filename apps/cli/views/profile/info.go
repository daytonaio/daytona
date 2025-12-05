// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package profile

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/util"
	"golang.org/x/term"
)

func RenderInfo(profile *config.Profile, activeProfileId *string, forceUnstyled bool) {
	var output string
	nameLabel := "Profile"

	output += "\n"
	output += getInfoLine(nameLabel, profile.Name)
	if activeProfileId != nil && *activeProfileId == profile.Id {
		output += getInfoLine("Status", "Active")
	}
	output += getInfoLine("ID", profile.Id) + "\n"
	output += getInfoLine("API URL", profile.Api.Url) + "\n"

	if profile.Api.Key != nil {
		output += getInfoLine("Auth", "API Key") + "\n"
	} else if profile.Api.Token != nil {
		output += getInfoLine("Auth", "Token") + "\n"
	} else {
		output += getInfoLine("Auth", "Not authenticated") + "\n"
	}

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}
	if terminalWidth < common.TUITableMinimumWidth || forceUnstyled {
		renderUnstyledInfo(output)
		return
	}

	output = common.GetStyledMainTitle("Profile Info") + "\n" + output

	renderTUIView(output, common.GetContainerBreakpointWidth(terminalWidth))
}

func renderUnstyledInfo(output string) {
	fmt.Println(output)
}

func renderTUIView(output string, width int) {
	output = lipgloss.NewStyle().PaddingLeft(3).Render(output)

	content := lipgloss.
		NewStyle().Width(width).
		Render(output)

	fmt.Println(content)
}

func getInfoLine(key, value string) string {
	return util.PropertyNameStyle.Render(fmt.Sprintf("%-*s", util.PropertyNameWidth, key)) + util.PropertyValueStyle.Render(value) + "\n"
}
