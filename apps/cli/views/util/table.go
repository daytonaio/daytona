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
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/daytonaio/daytona/cli/views/common"
	"golang.org/x/term"
)

var AdditionalPropertyPadding = "  "

// Left border, BaseTableStyle padding left, additional padding for target name and target config, BaseTableStyle padding right, BaseCellStyle padding right, right border
var RowWhiteSpace = 1 + 4 + len(AdditionalPropertyPadding)*2 + 4 + 4 + 1
var ArbitrarySpace = 10

// ansiRegex matches the full ANSI escape family, not just CSI color codes:
//   - OSC (Operating System Command) — terminal title, OSC-52 clipboard, OSC-8 hyperlinks
//   - DCS (Device Control String)
//   - CSI (Control Sequence Introducer) — colors, cursor movement, etc.
//   - Charset-designation escapes (ESC ( F, ESC ) F, etc.)
//   - Plain 2-byte ESC sequences (excluding [, ], P which introduce CSI/OSC/DCS)
//   - Stray bare ESC bytes as a last-resort fallback
//   - C0 control bytes and DEL (except \t \n \r — handled by tsvUnsafe below)
//
// A narrow CSI-only strip is unsafe here because user-controlled strings
// (sandbox names, label values) flow into output that may later be rendered
// in a real terminal. OSC-8 hyperlinks enable phishing; OSC-52 enables
// clipboard hijack; DCS can elicit terminal responses that the shell reads
// back as input. Order matters: Go's regexp picks leftmost-first among
// alternatives, so longer/more-specific patterns come first.
var ansiRegex = regexp.MustCompile(
	`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)` + // OSC, terminated by BEL or ST
		`|\x1bP[^\x1b]*\x1b\\` + // DCS, ST-terminated (before 2-byte ESC so \x1bP matches DCS)
		`|\x1b\[[0-?]*[ -/]*[@-~]` + // CSI: intro + params + intermediates + final
		`|\x1b[()*+\-./][@-~]` + // Charset designation: ESC <intermediate> <final>
		`|\x1b[@A-OQ-Z\\^_]` + // Plain 2-byte ESC (excludes [, ], P)
		`|\x1b` + // Stray bare ESC — drop alone
		`|[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`, // C0 / DEL (keeps \t \n \r — handled by tsvUnsafe)
)

// tsvUnsafe replaces row-/field-delimiter characters that would corrupt the
// TSV stream. The replacement choices match kubectl's convention (drop, don't
// quote) so downstream `awk -F'\t'` and `cut -f` parsers see a stable row
// count regardless of upstream field contents.
var tsvUnsafe = strings.NewReplacer("\t", " ", "\n", " ", "\r", " ")

// StripANSI removes ANSI escape sequences from s. See ansiRegex doc for the
// covered families.
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// SanitizeTSV prepares a string for emission as a TSV cell value. It strips
// ANSI escapes (preventing terminal injection when the output is later
// rendered) and replaces row/field delimiters (preventing row injection).
// Use this at every TSV emission site that handles user-controlled data.
func SanitizeTSV(s string) string {
	return tsvUnsafe.Replace(StripANSI(s))
}

// Gets the table view string or falls back to an unstyled view for lower terminal widths
func GetTableView(data [][]string, headers []string, activeOrganizationName *string, fallbackRender func()) string {
	if internal.FormatFlag == "tsv" {
		// Headers and the active-org banner are intentionally dropped in TSV
		// mode (kubectl-style): scripts grep/awk on data rows; a header row
		// would force them to `tail -n +2`.
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
			cells[i] = strings.TrimSpace(SanitizeTSV(cell))
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
