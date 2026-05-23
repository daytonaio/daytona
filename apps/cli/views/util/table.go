// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	commoncmd "github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/views/common"
	"golang.org/x/term"
)

var AdditionalPropertyPadding = "  "

// Left border, BaseTableStyle padding left, additional padding for target name and target config, BaseTableStyle padding right, BaseCellStyle padding right, right border
var RowWhiteSpace = 1 + 4 + len(AdditionalPropertyPadding)*2 + 4 + 4 + 1
var ArbitrarySpace = 10

var ansiRegex = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

// StripANSI removes ANSI escape sequences from s.
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// Gets the table view string or falls back to an unstyled view for lower terminal widths
func GetTableView(data [][]string, headers []string, activeOrganizationName *string, fallbackRender func()) string {
	if commoncmd.FormatFlag == "tsv" {
		return renderTSV(data)
	}

	re := lipgloss.NewRenderer(os.Stdout)

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(data)
		return ""
	}

	breakpointWidth := common.GetContainerBreakpointWidth(terminalWidth)

	minWidth := getMinimumWidth(data)

	if breakpointWidth == 0 || minWidth > breakpointWidth {
		fallbackRender()
		return ""
	}

	t := table.New().
		Headers(headers...).
		Rows(data...).
		BorderStyle(re.NewStyle().Foreground(common.LightGray)).
		BorderRow(false).BorderColumn(false).BorderLeft(false).BorderRight(false).BorderTop(false).BorderBottom(false).
		StyleFunc(func(_, _ int) lipgloss.Style {
			return common.BaseCellStyle
		}).Width(breakpointWidth - 2*common.BaseTableStyleHorizontalPadding - 1)

	table := t.String()

	if activeOrganizationName != nil {
		activeOrgMessage := common.GetInfoMessage(fmt.Sprintf("Active organization: %s", *activeOrganizationName))
		rightAlignedStyle := lipgloss.NewStyle().Width(breakpointWidth - 2*common.BaseTableStyleHorizontalPadding - 1).Align(lipgloss.Right)
		table += "\n" + rightAlignedStyle.Render(activeOrgMessage)
	}

	return common.BaseTableStyle.Render(table)
}

func getMinimumWidth(data [][]string) int {
	width := 0
	widestRow := 0
	for _, row := range data {
		for _, cell := range row {
			width += longestLineLength(StripANSI(cell))
			if width > widestRow {
				widestRow = width
			}
		}
		width = 0
	}
	return widestRow
}

// renderTSV emits the rows as literal tab-separated values, stripping any
// ANSI bytes the upstream styling layer may have applied. Headers are
// intentionally omitted (kubectl-style) so scripts like
// `daytona sandbox list | awk '{print $1}'` work without skipping a leading row.
func renderTSV(data [][]string) string {
	var b strings.Builder
	for _, row := range data {
		cells := make([]string, len(row))
		for i, cell := range row {
			cells[i] = strings.TrimSpace(StripANSI(cell))
		}
		b.WriteString(strings.Join(cells, "\t"))
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

// Returns the length of the longest line in a string
func longestLineLength(input string) int {
	lines := strings.Split(input, "\n")
	maxLength := 0

	for _, line := range lines {
		if len(line) > maxLength {
			maxLength = len(line)
		}
	}

	return maxLength
}
