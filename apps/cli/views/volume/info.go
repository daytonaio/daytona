// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/apiclient"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/util"
	"golang.org/x/term"
)

func RenderInfo(volume *apiclient.VolumeDto, forceUnstyled bool) {
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

func getStateLabel(state apiclient.VolumeState) string {
	switch state {
	case apiclient.VOLUMESTATE_PENDING_CREATE:
		return common.CreatingStyle.Render("PENDING CREATE")
	case apiclient.VOLUMESTATE_CREATING:
		return common.CreatingStyle.Render("CREATING")
	case apiclient.VOLUMESTATE_READY:
		return common.StartedStyle.Render("READY")
	case apiclient.VOLUMESTATE_PENDING_DELETE:
		return common.DeletedStyle.Render("PENDING DELETE")
	case apiclient.VOLUMESTATE_DELETING:
		return common.DeletedStyle.Render("DELETING")
	case apiclient.VOLUMESTATE_DELETED:
		return common.DeletedStyle.Render("DELETED")
	case apiclient.VOLUMESTATE_ERROR:
		return common.ErrorStyle.Render("ERROR")
	default:
		return common.UndefinedStyle.Render("/")
	}
}
