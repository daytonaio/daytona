// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	workspacetemplate_info "github.com/daytonaio/daytona/pkg/views/workspacetemplate/info"
	"golang.org/x/term"
)

const propertyNameWidth = 20

var propertyNameStyle = lipgloss.NewStyle().
	Foreground(views.LightGray)

var propertyValueStyle = lipgloss.NewStyle().
	Foreground(views.Light).
	Bold(true)

func Render(b *apiclient.Build, apiServerConfig *apiclient.ServerConfig, forceUnstyled bool) {
	var output string
	output += "\n\n"

	output += views.GetStyledMainTitle("Build Info") + "\n\n"

	output += getInfoLine("ID", b.Id) + "\n"

	output += getInfoLine("State", string(b.State)) + "\n"

	output += getInfoLine("Repository", b.Repository.Url) + "\n"

	if b.Image != nil {
		output += getInfoLine("Image", *b.Image) + "\n"
	}

	if b.User != nil {
		output += getInfoLine("User", *b.User) + "\n"
	}

	if workspacetemplate_info.GetLabelFromBuild(b.BuildConfig) != "" {
		workspaceDefaults := &views_util.WorkspaceTemplateDefaults{
			Image:     &apiServerConfig.DefaultWorkspaceImage,
			ImageUser: &apiServerConfig.DefaultWorkspaceUser,
		}

		_, buildChoice := views_util.GetWorkspaceBuildChoice(apiclient.CreateWorkspaceDTO{
			BuildConfig: b.BuildConfig,
		}, workspaceDefaults)
		output += getInfoLine("Build", buildChoice) + "\n"
	}

	if b.BuildConfig != nil && b.BuildConfig.Devcontainer != nil {
		output += getInfoLine("Devcontainer path", b.BuildConfig.Devcontainer.FilePath) + "\n"
	}

	output += getInfoLine("Prebuild ID", b.PrebuildId) + "\n"

	output += getInfoLine("Created", util.FormatTimestamp(b.CreatedAt)) + "\n"

	output += getInfoLine("Updated", util.FormatTimestamp(b.UpdatedAt)) + "\n"

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
