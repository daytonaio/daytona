// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var linkStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

var docsURL string = "https://www.daytona.io/docs/"

var DocsCmd = &cobra.Command{
	Use:     "docs",
	Short:   "Opens the Daytona documentation in your default browser.",
	Args:    cobra.NoArgs,
	Aliases: []string{"documentation", "doc"},
	Run: func(cmd *cobra.Command, args []string) {
		output := views.GetBoldedInfoMessage("Opening the Daytona documentation in your default browser. If opening fails, you can go to " + linkStyle.Render(docsURL) + " manually.")
		fmt.Println(output)
		err := browser.OpenURL(docsURL)
		if err != nil {
			log.Fatal(err)
		}
	},
}
