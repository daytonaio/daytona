// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	Green      = lipgloss.AdaptiveColor{Light: "#23cc71", Dark: "#23cc71"}
	Blue       = lipgloss.AdaptiveColor{Light: "#017ffe", Dark: "#017ffe"}
	DimmedBlue = lipgloss.AdaptiveColor{Light: "#3398fe", Dark: "#3398fe"}
	White      = lipgloss.AdaptiveColor{Light: "000", Dark: "fff"}
)

func ColorGrid(xSteps, ySteps int) [][]string {
	x0y0, _ := colorful.Hex("#F25D94")
	x1y0, _ := colorful.Hex("#EDFF82")
	x0y1, _ := colorful.Hex("#643AFF")
	x1y1, _ := colorful.Hex("#14F9D5")

	x0 := make([]colorful.Color, ySteps)
	for i := range x0 {
		x0[i] = x0y0.BlendLuv(x0y1, float64(i)/float64(ySteps))
	}

	x1 := make([]colorful.Color, ySteps)
	for i := range x1 {
		x1[i] = x1y0.BlendLuv(x1y1, float64(i)/float64(ySteps))
	}

	grid := make([][]string, ySteps)
	for x := 0; x < ySteps; x++ {
		y0 := x0[x]
		grid[x] = make([]string, xSteps)
		for y := 0; y < xSteps; y++ {
			grid[x][y] = y0.BlendLuv(x1[x], float64(y)/float64(xSteps)).Hex()
		}
	}

	return grid
}

func GetCustomTheme() *huh.Theme {

	newTheme := huh.ThemeCharm()

	newTheme.Blurred.Title = newTheme.Focused.Title

	b := &newTheme.Blurred
	b.FocusedButton.Background(Green)
	b.FocusedButton.Bold(true)
	b.TextInput.Prompt.Foreground(Green)
	b.TextInput.Cursor.Foreground(Green)
	b.SelectSelector.Foreground(Green)

	f := &newTheme.Focused
	f.Base = f.Base.BorderForeground(lipgloss.Color("fff"))
	f.Title.Foreground(Blue).Bold(true)
	f.FocusedButton.Bold(true)
	f.FocusedButton.Background(Green)
	f.TextInput.Prompt.Foreground(Green)
	f.TextInput.Cursor.Foreground(White)
	f.SelectSelector.Foreground(Green)

	return newTheme
}
