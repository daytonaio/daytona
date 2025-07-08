// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/util"
	"golang.org/x/term"
)

func RenderInfo(snapshot *apiclient.SnapshotDto, forceUnstyled bool) {
	var output string
	nameLabel := "Snapshot"

	output += "\n"
	output += getInfoLine(nameLabel, snapshot.Name) + "\n"
	output += getInfoLine("State", getStateLabel(snapshot.State)) + "\n"
	output += getInfoLine("Enabled", fmt.Sprintf("%v", snapshot.Enabled)) + "\n"

	if snapshot.Size.IsSet() {
		output += getInfoLine("Size", fmt.Sprintf("%.2f GB", *snapshot.Size.Get())) + "\n"
	} else {
		output += getInfoLine("Size", "-") + "\n"
	}
	output += getInfoLine("Created", util.GetTimeSinceLabel(snapshot.CreatedAt)) + "\n"

	output += getInfoLine("ID", snapshot.Id) + "\n"

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}
	if terminalWidth < common.TUITableMinimumWidth || forceUnstyled {
		renderUnstyledInfo(output)
		return
	}

	output = common.GetStyledMainTitle("Snapshot Info") + "\n" + output

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

func getStateLabel(state apiclient.SnapshotState) string {
	switch state {
	case apiclient.SNAPSHOTSTATE_PENDING:
		return common.CreatingStyle.Render("PENDING")
	case apiclient.SNAPSHOTSTATE_PULLING:
		return common.CreatingStyle.Render("PULLING SNAPSHOT")
	case apiclient.SNAPSHOTSTATE_VALIDATING:
		return common.CreatingStyle.Render("VALIDATING")
	case apiclient.SNAPSHOTSTATE_ACTIVE:
		return common.StartedStyle.Render("ACTIVE")
	case apiclient.SNAPSHOTSTATE_ERROR:
		return common.ErrorStyle.Render("ERROR")
	case apiclient.SNAPSHOTSTATE_BUILD_FAILED:
		return common.ErrorStyle.Render("BUILD FAILED")
	case apiclient.SNAPSHOTSTATE_REMOVING:
		return common.DeletedStyle.Render("REMOVING")
	default:
		return common.UndefinedStyle.Render("/")
	}
}
