// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "regexp"

var AdditionalPropertyPadding = "  "

// Left border, BaseTableStyle padding left, additional padding for workspace name and target, BaseTableStyle padding right, BaseCellStyle padding right, right border
var RowWhiteSpace = 1 + 4 + len(AdditionalPropertyPadding)*2 + 4 + 4 + 1
var ArbitrarySpace = 10

func GetTableMinimumWidth(data [][]string) int {
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
