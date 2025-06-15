// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/util"
	"github.com/daytonaio/daytona/daytonaapiclient"
	"golang.org/x/term"
)

func RenderInfo(sandbox *daytonaapiclient.Sandbox, forceUnstyled bool) {
	var output string

	output += "\n"

	output += getInfoLine("ID", sandbox.Id) + "\n"

	if sandbox.State != nil {
		output += getInfoLine("State", getStateLabel(*sandbox.State)) + "\n"
	}

	if sandbox.Snapshot != nil {
		output += getInfoLine("Snapshot", *sandbox.Snapshot) + "\n"
	}

	output += getInfoLine("Region", sandbox.Target) + "\n"

	if sandbox.Class != nil {
		output += getInfoLine("Class", *sandbox.Class) + "\n"
	}

	if sandbox.CreatedAt != nil {
		output += getInfoLine("Created", util.GetTimeSinceLabelFromString(*sandbox.CreatedAt)) + "\n"
	}

	if sandbox.UpdatedAt != nil {
		output += getInfoLine("Last Event", util.GetTimeSinceLabelFromString(*sandbox.UpdatedAt)) + "\n"
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

	output = common.GetStyledMainTitle("Sandbox Info") + "\n" + output

	if len(sandbox.Labels) > 0 {
		labels := ""
		i := 0
		for k, v := range sandbox.Labels {
			label := fmt.Sprintf("%s=%s\n", k, v)
			if i == 0 {
				labels += label + "\n"
			} else {
				labels += getInfoLine("", fmt.Sprintf("%s=%s\n", k, v))
			}
			i++
		}
		labels = strings.TrimSuffix(labels, "\n")
		output += "\n" + strings.TrimSuffix(getInfoLine("Labels", labels), "\n")
	}

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

func getStateLabel(state daytonaapiclient.SandboxState) string {
	switch state {
	case daytonaapiclient.SANDBOXSTATE_CREATING:
		return common.CreatingStyle.Render("CREATING")
	case daytonaapiclient.SANDBOXSTATE_RESTORING:
		return common.CreatingStyle.Render("RESTORING")
	case daytonaapiclient.SANDBOXSTATE_DESTROYED:
		return common.DeletedStyle.Render("DESTROYED")
	case daytonaapiclient.SANDBOXSTATE_DESTROYING:
		return common.DeletedStyle.Render("DESTROYING")
	case daytonaapiclient.SANDBOXSTATE_STARTED:
		return common.StartedStyle.Render("STARTED")
	case daytonaapiclient.SANDBOXSTATE_STOPPED:
		return common.StoppedStyle.Render("STOPPED")
	case daytonaapiclient.SANDBOXSTATE_STARTING:
		return common.StartingStyle.Render("STARTING")
	case daytonaapiclient.SANDBOXSTATE_STOPPING:
		return common.StoppingStyle.Render("STOPPING")
	case daytonaapiclient.SANDBOXSTATE_PULLING_SNAPSHOT:
		return common.CreatingStyle.Render("PULLING SNAPSHOT")
	case daytonaapiclient.SANDBOXSTATE_ARCHIVING:
		return common.CreatingStyle.Render("ARCHIVING")
	case daytonaapiclient.SANDBOXSTATE_ARCHIVED:
		return common.StoppedStyle.Render("ARCHIVED")
	case daytonaapiclient.SANDBOXSTATE_ERROR:
		return common.ErrorStyle.Render("ERROR")
	case daytonaapiclient.SANDBOXSTATE_BUILD_FAILED:
		return common.ErrorStyle.Render("BUILD FAILED")
	case daytonaapiclient.SANDBOXSTATE_UNKNOWN:
		return common.UndefinedStyle.Render("UNKNOWN")
	default:
		return common.UndefinedStyle.Render("/")
	}
}
