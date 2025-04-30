// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona-ai-saas/cli/views/common"
	"github.com/daytonaio/daytona-ai-saas/cli/views/util"
	"github.com/daytonaio/daytona-ai-saas/daytonaapiclient"
	"golang.org/x/term"
)

func RenderInfo(sandbox *daytonaapiclient.Workspace, forceUnstyled bool) {
	var output string

	output += "\n"

	output += getInfoLine("ID", sandbox.Id) + "\n"

	if sandbox.State != nil {
		output += getInfoLine("State", getStateLabel(*sandbox.State)) + "\n"
	}

	if sandbox.Image != nil {
		output += getInfoLine("Image", *sandbox.Image) + "\n"
	}

	providerMetadataString := sandbox.Info.GetProviderMetadata()

	var providerMetadata providerMetadata

	err := json.Unmarshal([]byte(providerMetadataString), &providerMetadata)
	if err == nil {
		output += getInfoLine("Region", providerMetadata.Region) + "\n"
		output += getInfoLine("Class", providerMetadata.Class) + "\n"
		output += getInfoLine("Last Event", util.GetTimeSinceLabelFromString(providerMetadata.UpdatedAt)) + "\n"
	}

	if sandbox.Info != nil {
		output += getInfoLine("Created", util.GetTimeSinceLabelFromString(sandbox.Info.Created)) + "\n"
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

type providerMetadata struct {
	Region    string `json:"region"`
	Class     string `json:"class"`
	UpdatedAt string `json:"updatedAt"`
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

func getStateLabel(state daytonaapiclient.WorkspaceState) string {
	switch state {
	case daytonaapiclient.WORKSPACESTATE_CREATING:
		return common.CreatingStyle.Render("CREATING")
	case daytonaapiclient.WORKSPACESTATE_RESTORING:
		return common.CreatingStyle.Render("RESTORING")
	case daytonaapiclient.WORKSPACESTATE_DESTROYED:
		return common.DeletedStyle.Render("DESTROYED")
	case daytonaapiclient.WORKSPACESTATE_DESTROYING:
		return common.DeletedStyle.Render("DESTROYING")
	case daytonaapiclient.WORKSPACESTATE_STARTED:
		return common.StartedStyle.Render("STARTED")
	case daytonaapiclient.WORKSPACESTATE_STOPPED:
		return common.StoppedStyle.Render("STOPPED")
	case daytonaapiclient.WORKSPACESTATE_STARTING:
		return common.StartingStyle.Render("STARTING")
	case daytonaapiclient.WORKSPACESTATE_STOPPING:
		return common.StoppingStyle.Render("STOPPING")
	case daytonaapiclient.WORKSPACESTATE_PULLING_IMAGE:
		return common.CreatingStyle.Render("PULLING IMAGE")
	case daytonaapiclient.WORKSPACESTATE_ARCHIVING:
		return common.CreatingStyle.Render("ARCHIVING")
	case daytonaapiclient.WORKSPACESTATE_ARCHIVED:
		return common.StoppedStyle.Render("ARCHIVED")
	case daytonaapiclient.WORKSPACESTATE_ERROR:
		return common.ErrorStyle.Render("ERROR")
	case daytonaapiclient.WORKSPACESTATE_UNKNOWN:
		return common.UndefinedStyle.Render("UNKNOWN")
	default:
		return common.UndefinedStyle.Render("/")
	}
}
