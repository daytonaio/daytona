// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var wrappedText strings.Builder
	var line strings.Builder

	words := strings.Fields(text)

	for _, word := range words {
		if utf8.RuneCountInString(line.String())+utf8.RuneCountInString(word)+1 > width-2 {
			wrappedText.WriteString(line.String() + "\n")
			line.Reset()
		}

		if line.Len() > 0 {
			line.WriteString(" ")
		}
		line.WriteString(word)
	}

	if line.Len() > 0 {
		wrappedText.WriteString(line.String())
	}

	return wrappedText.String()
}

func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	const maxWidth = 160
	if err != nil {
		return maxWidth
	}
	return width
}
func GetTerminalHeight() int {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	const maxHeight = 50
	if err != nil {
		return maxHeight
	}
	return height
}
