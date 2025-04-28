// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona-ai-saas/cli/views/common"
	"github.com/daytonaio/daytona-ai-saas/cli/views/util"
	"github.com/daytonaio/daytona-ai-saas/daytonaapiclient"
	"golang.org/x/term"
)

func RenderInfo(volume *daytonaapiclient.VolumeDto, forceUnstyled bool) {
	var output string
	nameLabel := "Volume"

	output += "\n"
	output += getInfoLine(nameLabel, volume.Name) + "\n"
	output += getInfoLine("ID", volume.Id) + "\n"
	output += getInfoLine("State", getStateLabel(volume.State)) + "\n"

	output += getInfoLine("Created", util.GetTimeSinceLabelFromString(volume.CreatedAt)) + "\n"

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}
	if terminalWidth < common.TUITableMinimumWidth || forceUnstyled {
		renderUnstyledInfo(output)
		return
	}

	output = common.GetStyledMainTitle("Volume Info") + "\n" + output

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

func getStateLabel(state string) string {
	switch state {
	case "creating":
		return common.CreatingStyle.Render("CREATING")
	case "available":
		return common.StartedStyle.Render("AVAILABLE")
	case "deleting":
		return common.DeletedStyle.Render("DELETING")
	case "error":
		return common.ErrorStyle.Render("ERROR")
	default:
		return common.UndefinedStyle.Render("/")
	}
}
