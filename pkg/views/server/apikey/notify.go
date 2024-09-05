// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

var minimumLayoutWidth = 80

func Render(key, apiUrl string) {
	var output string

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println("error: Unable to get terminal size")
		return
	}

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Generated API key: "), key) + "\n\n"

	output += "Make sure to copy it as you will not be able to see it again." + "\n\n"

	output += views.SeparatorString + "\n\n"

	output += "You can connect to the Daytona Server from a client machine by running:"

	formattedCommand := lipgloss.NewStyle().Foreground(views.Green).Render(fmt.Sprintf("daytona profile add -a \\\n%s \\\n-k %s", apiUrl, key))
	command := fmt.Sprintf("daytona profile add -a %s -k %s", apiUrl, key)

	clipboardMessage := copyToClipboard(command)

	if terminalWidth >= minimumLayoutWidth {
		output += "\n\n" + formattedCommand
		output += "\n\n" + clipboardMessage
		views.RenderContainerLayout(views.GetInfoMessage(output))
	} else {
		views.RenderContainerLayout(views.GetInfoMessage(output))
		fmt.Println(command + "\n\n")
		fmt.Println(clipboardMessage)
	}
}

func copyToClipboard(command string) string {
	err := clipboard.WriteAll(command)
	if err != nil {
		return "Error: Unable to copy the command to the clipboard. Please copy it manually."
	}
	return "The command has been successfully copied to your clipboard."
}
