// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"golang.org/x/term"
)

const propertyNameWidth = 16

var propertyNameStyle = lipgloss.NewStyle().
	Foreground(views.LightGray)

var propertyValueStyle = lipgloss.NewStyle().
	Foreground(views.Light).
	Bold(true)

func Render(target *apiclient.TargetDTO, forceUnstyled bool) {
	var output string
	nameLabel := "Name"

	output += "\n"
	output += getInfoLine(nameLabel, target.Name) + "\n"

	output += getInfoLine("ID", target.Id) + "\n"

	providerLabel := target.TargetConfig.ProviderInfo.Name
	if target.TargetConfig.ProviderInfo.Label != nil {
		providerLabel = *target.TargetConfig.ProviderInfo.Label
	}

	output += getInfoLine("Provider", providerLabel) + "\n"

	output += getInfoLine("Runner", target.TargetConfig.ProviderInfo.RunnerName) + "\n"

	if target.Default {
		output += getInfoLine("Default", "Yes") + "\n"
	}

	output += getInfoLineState("State", target.State, target.Metadata) + "\n"
	if target.State.Error != nil {
		output += getInfoLine("Error", *target.State.Error) + "\n"
	}

	output += getInfoLine("# Workspaces", fmt.Sprint(len(target.Workspaces))) + "\n"

	output += getInfoLine("Options", target.TargetConfig.Options) + "\n"

	if target.ProviderMetadata != nil {
		output += getInfoLine("Metadata", *target.ProviderMetadata) + "\n"
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

func getInfoLineState(key string, state apiclient.ResourceState, metadata *apiclient.TargetMetadata) string {
	stateLabel := views.GetStateLabel(state.Name)

	if metadata != nil {
		views_util.CheckAndAppendTimeLabel(&stateLabel, state, metadata.Uptime)
	}
	return propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key)) + stateLabel + propertyValueStyle.Foreground(views.Light).Render("\n")
}
