// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package organization

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/util"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"golang.org/x/term"
)

func RenderInfo(organization *apiclient.Organization, forceUnstyled bool) {
	if internal.FormatFlag == "tsv" {
		renderTSVInfo(os.Stdout, organization)
		return
	}

	var output string
	nameLabel := "Organization"

	output += "\n"
	output += getInfoLine(nameLabel, organization.Name) + "\n"
	output += getInfoLine("Created", util.GetTimeSinceLabel(organization.CreatedAt)) + "\n"
	output += getInfoLine("ID", organization.Id) + "\n"

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}

	if terminalWidth < common.TUITableMinimumWidth || forceUnstyled {
		renderUnstyledInfo(output)
		return
	}

	output = common.GetStyledMainTitle("Organization Info") + "\n" + output

	renderTUIView(output, common.GetContainerBreakpointWidth(terminalWidth))
}

func renderUnstyledInfo(output string) {
	fmt.Println(output)
}

func renderTSVInfo(w io.Writer, o *apiclient.Organization) {
	fmt.Fprintf(w, "organization\t%s\n", util.SanitizeTSV(o.Name))
	fmt.Fprintf(w, "created\t%s\n", o.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "id\t%s\n", util.SanitizeTSV(o.Id))
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
