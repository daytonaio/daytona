// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package image

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/util"
	"github.com/daytonaio/daytona/daytonaapiclient"
	"golang.org/x/term"
)

func RenderInfo(image *daytonaapiclient.ImageDto, forceUnstyled bool) {
	var output string
	nameLabel := "Image"

	output += "\n"
	output += getInfoLine(nameLabel, image.Name) + "\n"
	output += getInfoLine("State", getStateLabel(image.State)) + "\n"
	output += getInfoLine("Enabled", fmt.Sprintf("%v", image.Enabled)) + "\n"

	if image.Size.IsSet() {
		output += getInfoLine("Size", fmt.Sprintf("%.2f GB", *image.Size.Get())) + "\n"
	} else {
		output += getInfoLine("Size", "-") + "\n"
	}
	output += getInfoLine("Created", util.GetTimeSinceLabel(image.CreatedAt)) + "\n"

	output += getInfoLine("ID", image.Id) + "\n"

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}
	if terminalWidth < common.TUITableMinimumWidth || forceUnstyled {
		renderUnstyledInfo(output)
		return
	}

	output = common.GetStyledMainTitle("Image Info") + "\n" + output

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

func getStateLabel(state daytonaapiclient.ImageState) string {
	switch state {
	case daytonaapiclient.IMAGESTATE_PENDING:
		return common.CreatingStyle.Render("PENDING")
	case daytonaapiclient.IMAGESTATE_PULLING_IMAGE:
		return common.CreatingStyle.Render("PULLING IMAGE")
	case daytonaapiclient.IMAGESTATE_VALIDATING:
		return common.CreatingStyle.Render("VALIDATING")
	case daytonaapiclient.IMAGESTATE_ACTIVE:
		return common.StartedStyle.Render("ACTIVE")
	case daytonaapiclient.IMAGESTATE_ERROR:
		return common.ErrorStyle.Render("ERROR")
	case daytonaapiclient.IMAGESTATE_REMOVING:
		return common.DeletedStyle.Render("REMOVING")
	default:
		return common.UndefinedStyle.Render("/")
	}
}
