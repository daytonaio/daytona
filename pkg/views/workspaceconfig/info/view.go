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
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"golang.org/x/term"
)

const propertyNameWidth = 20

var propertyNameStyle = lipgloss.NewStyle().
	Foreground(views.LightGray)

var propertyValueStyle = lipgloss.NewStyle().
	Foreground(views.Light).
	Bold(true)

var prebuildDetailStyle = propertyNameStyle

func Render(workspaceConfig *apiclient.WorkspaceConfig, apiServerConfig *apiclient.ServerConfig, forceUnstyled bool) {
	var output string
	output += "\n\n"

	output += views.GetStyledMainTitle("Workspace Config Info") + "\n\n"

	output += getInfoLine("Name", workspaceConfig.Name) + "\n"

	output += getInfoLine("Repository", workspaceConfig.RepositoryUrl) + "\n"

	gitProviderConfig := "/"
	if workspaceConfig.GitProviderConfigId != nil {
		gitProviderConfig = *workspaceConfig.GitProviderConfigId
	}

	output += getInfoLine("Git Provider ID", gitProviderConfig) + "\n"

	if workspaceConfig.Default {
		output += getInfoLine("Default", "Yes") + "\n"
	}

	if GetLabelFromBuild(workspaceConfig.BuildConfig) != "" {
		workspaceDefaults := &views_util.WorkspaceConfigDefaults{
			Image:     &apiServerConfig.DefaultWorkspaceImage,
			ImageUser: &apiServerConfig.DefaultWorkspaceUser,
		}

		createWorkspaceDto := apiclient.CreateWorkspaceDTO{
			BuildConfig: workspaceConfig.BuildConfig,
		}
		_, buildChoice := views_util.GetWorkspaceBuildChoice(createWorkspaceDto, workspaceDefaults)
		output += getInfoLine("Build", buildChoice) + "\n"
	}

	if workspaceConfig.Image != "" {
		output += getInfoLine("Image", workspaceConfig.Image) + "\n"
	}

	if workspaceConfig.User != "" {
		output += getInfoLine("User", workspaceConfig.User) + "\n"
	}

	if workspaceConfig.BuildConfig != nil && workspaceConfig.BuildConfig.Devcontainer != nil {
		output += getInfoLine("Devcontainer path", workspaceConfig.BuildConfig.Devcontainer.FilePath) + "\n"
	}

	prebuildCount := len(workspaceConfig.Prebuilds)

	if prebuildCount > 0 {
		if prebuildCount == 1 {
			output += getInfoLine("Prebuild: ", getPrebuildLine(workspaceConfig.Prebuilds[0], nil)) + "\n"
		} else {
			output += getInfoLine("Prebuilds: ", "") + "\n"
			for i, prebuild := range workspaceConfig.Prebuilds {
				if len(workspaceConfig.Prebuilds) != 1 {
					output += getPrebuildLine(prebuild, util.Pointer(i+1)) + "\n"
				} else {
					output += getPrebuildLine(prebuild, nil) + "\n"
				}
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

func getPrebuildLine(prebuild apiclient.PrebuildConfig, order *int) string {
	var line string
	if order != nil {
		line += propertyNameStyle.Render(fmt.Sprintf("%s#%d%s", strings.Repeat(" ", 3), *order, strings.Repeat(" ", 2)))
	}

	line += propertyValueStyle.Render(views.GetBranchNameLabel(prebuild.Branch))
	line += prebuildDetailStyle.Render(fmt.Sprintf(" - every %d commits - retention: %d builds", prebuild.CommitInterval, prebuild.Retention))

	if order != nil {
		line += "\n"
	}

	return line
}

func GetLabelFromBuild(build *apiclient.BuildConfig) string {
	if build == nil {
		return "Automatic"
	}

	if build.Devcontainer != nil {
		return fmt.Sprintf("Devcontainer (%s)", build.Devcontainer.FilePath)
	}

	return ""
}
