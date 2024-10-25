// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"os"
	"regexp"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

var AdditionalPropertyPadding = "  "

// Left border, BaseTableStyle padding left, additional padding for target name and target config, BaseTableStyle padding right, BaseCellStyle padding right, right border
var RowWhiteSpace = 1 + 4 + len(AdditionalPropertyPadding)*2 + 4 + 4 + 1
var ArbitrarySpace = 10

// Gets the table view string or falls back to an unstyled view for lower terminal widths
func GetTableView(data [][]string, headers []string, footer *string, fallbackRender func()) string {
	re := lipgloss.NewRenderer(os.Stdout)

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(data)
		return ""
	}

	breakpointWidth := views.GetContainerBreakpointWidth(terminalWidth)

	minWidth := getMinimumWidth(data)

	if breakpointWidth == 0 || minWidth > breakpointWidth {
		fallbackRender()
		return ""
	}

	t := table.New().
		Headers(headers...).
		Rows(data...).
		BorderStyle(re.NewStyle().Foreground(views.LightGray)).
		BorderRow(false).BorderColumn(false).BorderLeft(false).BorderRight(false).BorderTop(false).BorderBottom(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return views.TableHeaderStyle
			}
			return views.BaseCellStyle
		}).Width(breakpointWidth - 2*views.BaseTableStyleHorizontalPadding - 1)

	table := t.String()

	if footer != nil {
		table += "\n" + *footer
	}

	return views.BaseTableStyle.Render(table)
}

func getMinimumWidth(data [][]string) int {
	width := 0
	widestRow := 0
	for _, row := range data {
		for _, cell := range row {
			// Remove ANSI escape codes
			regex := regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")
			strippedCell := regex.ReplaceAllString(cell, "")
			width += len(strippedCell)
			if width > widestRow {
				widestRow = width
			}
		}
		width = 0
	}
	return widestRow
}
