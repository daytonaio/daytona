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

const minTUIWidth = 80
const maxTUIWidth = 140

func Render(apiPort uint32, frpcUrl string, isDaemonMode bool) {
	output := "\n"
	output += views.GetStyledMainTitle("Daytona") + "\n\n"
	output += fmt.Sprintf("## Daytona Server is running on port: %d\n\n", apiPort)
	output += views.GetSeparatorString() + "\n\n"
	output += "You may now begin developing locally"
	output += "\n"

	// output += fmt.Sprintf("If you want to connect to the server remotely:\n\n")

	// output += fmt.Sprintf("1. Create an API key on this machine: ")
	// output += lipgloss.NewStyle().Foreground(views.Green).Render("daytona server api-key new") + "\n"
	// output += fmt.Sprintf("2. Add a profile on the client machine: \n\t")
	// output += lipgloss.NewStyle().Foreground(views.Green).Render(fmt.Sprintf("daytona profile add -a %s -k API_KEY", m.frpcUrl)) + "\n\n"

	// output += views.GetSeparatorString() + "\n\n"

	// output += "Press Enter to create an API key and copy the complete client command to clipboard automatically"

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || terminalWidth < minTUIWidth {
		fmt.Println(output)
		return
	}
	width := terminalWidth - 20
	if width > maxTUIWidth {
		width = maxTUIWidth
	}

	output = lipgloss.NewStyle().PaddingLeft(3).Render(output)

	if !isDaemonMode {
		output = "\n" + output
	}

	output = lipgloss.
		NewStyle().
		BorderForeground(views.LightGray).
		Border(lipgloss.RoundedBorder()).Width(width).
		Render(output) + "\n"

	fmt.Println(output)
}
