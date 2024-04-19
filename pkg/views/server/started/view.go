// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package started

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

func Render(apiPort uint32, frpcUrl string, isDaemonMode bool) {
	output := "\n"
	output += views.GetStyledMainTitle("Daytona") + "\n\n"
	output += fmt.Sprintf("## Daytona Server is running on port: %d\n\n", apiPort)
	output += views.SeparatorString + "\n\n"
	output += "You may now begin developing locally"
	output += "\n"

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}

	output = lipgloss.NewStyle().PaddingLeft(3).Render(output)

	output = lipgloss.
		NewStyle().
		BorderForeground(views.LightGray).
		Border(lipgloss.RoundedBorder()).Width(views.GetContainerBreakpointWidth(terminalWidth)).
		Render(output) + "\n"

	if !isDaemonMode {
		output = "\n" + output
	}

	fmt.Println(output)
}
