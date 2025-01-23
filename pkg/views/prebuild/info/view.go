// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

const propertyNameWidth = 20

var propertyNameStyle = lipgloss.NewStyle().
	Foreground(views.LightGray)

var propertyValueStyle = lipgloss.NewStyle().
	Foreground(views.Light).
	Bold(true)

func Render(prebuild *apiclient.PrebuildDTO, forceUnstyled bool) {
	var output string
	output += "\n\n"

	output += views.GetStyledMainTitle("Prebuild Configuration Info") + "\n\n"

	output += getInfoLine("ID", prebuild.Id) + "\n"

	output += getInfoLine("Workspace template", prebuild.WorkspaceTemplateName) + "\n"

	output += getInfoLine("Branch", views.GetBranchNameLabel(prebuild.Branch)) + "\n"

	if prebuild.CommitInterval != nil {
		output += getInfoLine("Commit interval", fmt.Sprint(*prebuild.CommitInterval)) + "\n"
	}

	output += getInfoLine("Build retention", fmt.Sprint(prebuild.Retention)) + "\n"

	triggerFileCount := len(prebuild.TriggerFiles)

	if triggerFileCount > 0 {
		if triggerFileCount == 1 {
			output += getInfoLine("Trigger file: ", getTriggerFileLine(prebuild.TriggerFiles[0], nil)) + "\n"
		} else {
			output += getInfoLine("Trigger files:", "") + "\n"
			for i, triggerFile := range prebuild.TriggerFiles {
				output += getTriggerFileLine(triggerFile, util.Pointer(i+1)) + "\n"
			}
		}
	}

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}
	if terminalWidth < views.TUITableMinimumWidth || forceUnstyled {
		renderUnstyledInfo(output)
		return
	}

	renderTUIView(output, views.GetContainerBreakpointWidth(terminalWidth))
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
	return propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key)) + propertyValueStyle.Render(value) + "\n"
}

func getTriggerFileLine(file string, order *int) string {
	var line string
	if order != nil {
		line += propertyNameStyle.Render(fmt.Sprintf("%s#%d%s", strings.Repeat(" ", 3), *order, strings.Repeat(" ", 2)))
	}
	return line + propertyValueStyle.Render(file)
}
