// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

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

func RenderInfo(snapshot *apiclient.SnapshotDto, forceUnstyled bool) {
	if internal.FormatFlag == "tsv" {
		renderTSVInfo(os.Stdout, snapshot)
		return
	}

	var output string
	nameLabel := "Snapshot"

	output += "\n"
	output += getInfoLine(nameLabel, snapshot.Name) + "\n"
	output += getInfoLine("State", getStateLabel(snapshot.State)) + "\n"

	if size := snapshot.Size.Get(); size != nil {
		output += getInfoLine("Size", fmt.Sprintf("%.2f GB", *size)) + "\n"
	} else {
		output += getInfoLine("Size", "-") + "\n"
	}
	output += getInfoLine("Created", util.GetTimeSinceLabel(snapshot.CreatedAt)) + "\n"

	output += getInfoLine("ID", snapshot.Id) + "\n"

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}
	if terminalWidth < common.TUITableMinimumWidth || forceUnstyled {
		renderUnstyledInfo(output)
		return
	}

	output = common.GetStyledMainTitle("Snapshot Info") + "\n" + output

	renderTUIView(output, common.GetContainerBreakpointWidth(terminalWidth))
}

func renderUnstyledInfo(output string) {
	fmt.Println(output)
}

func renderTSVInfo(w io.Writer, s *apiclient.SnapshotDto) {
	fmt.Fprintf(w, "snapshot\t%s\n", util.SanitizeTSV(s.Name))
	fmt.Fprintf(w, "state\t%s\n", util.SanitizeTSV(string(s.State)))
	if size := s.Size.Get(); size != nil {
		fmt.Fprintf(w, "size_gb\t%.2f\n", *size)
	}
	fmt.Fprintf(w, "created\t%s\n", s.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "id\t%s\n", util.SanitizeTSV(s.Id))
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

func getStateLabel(state apiclient.SnapshotState) string {
	switch state {
	case apiclient.SNAPSHOTSTATE_PENDING:
		return common.CreatingStyle.Render("PENDING")
	case apiclient.SNAPSHOTSTATE_PULLING:
		return common.CreatingStyle.Render("PULLING SNAPSHOT")
	case apiclient.SNAPSHOTSTATE_ACTIVE:
		return common.StartedStyle.Render("ACTIVE")
	case apiclient.SNAPSHOTSTATE_ERROR:
		return common.ErrorStyle.Render("ERROR")
	case apiclient.SNAPSHOTSTATE_BUILD_FAILED:
		return common.ErrorStyle.Render("BUILD FAILED")
	case apiclient.SNAPSHOTSTATE_REMOVING:
		return common.DeletedStyle.Render("REMOVING")
	default:
		return common.UndefinedStyle.Render("/")
	}
}
