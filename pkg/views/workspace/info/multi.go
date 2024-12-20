// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

func RenderMulti(workspaces []apiclient.WorkspaceDTO, ide string, forceUnstyled bool) {
	var output string

	output += "\n"

	output += getInfoLine("Using Target", workspaces[0].Target.Name) + "\n"
	output += getInfoLine("Target ID", workspaces[0].TargetId) + "\n"

	output += getInfoLine("Editor", ide) + "\n"

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}
	if terminalWidth < views.TUITableMinimumWidth || forceUnstyled {
		renderUnstyledInfo(output)
		return
	}

	output += fmt.Sprintf("%s\n\n", views.SeparatorString)
	for index, workspace := range workspaces {
		output += getInfoLine(fmt.Sprintf("Workspace #%d", index+1), fmt.Sprintf("%s (%s)", workspace.Name, workspace.Id)) + "\n"
		output += getWorkspaceDataOutput(&workspace, true)
		if index < len(workspaces)-1 {
			output += fmt.Sprintf("\n%s\n\n", views.SeparatorString)
		}
	}

	renderTUIView(output, views.GetContainerBreakpointWidth(terminalWidth), true)
}
