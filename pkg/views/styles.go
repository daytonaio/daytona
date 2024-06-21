// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	Green       = lipgloss.AdaptiveColor{Light: "#23cc71", Dark: "#23cc71"}
	Blue        = lipgloss.AdaptiveColor{Light: "#017ffe", Dark: "#017ffe"}
	Yellow      = lipgloss.AdaptiveColor{Light: "#d4ed2d", Dark: "#d4ed2d"}
	Cyan        = lipgloss.AdaptiveColor{Light: "#3ef7e5", Dark: "#3ef7e5"}
	DimmedGreen = lipgloss.AdaptiveColor{Light: "#7be0a9", Dark: "#7be0a9"}
	Orange      = lipgloss.AdaptiveColor{Light: "#e3881b", Dark: "#e3881b"}
	Light       = lipgloss.AdaptiveColor{Light: "#000", Dark: "#fff"}
	Dark        = lipgloss.AdaptiveColor{Light: "#fff", Dark: "#000"}
	Gray        = lipgloss.AdaptiveColor{Light: "243", Dark: "243"}
	LightGray   = lipgloss.AdaptiveColor{Light: "#828282", Dark: "#828282"}
)

var (
	BaseTableStyleHorizontalPadding = 4
	BaseTableStyle                  = lipgloss.NewStyle().
					PaddingLeft(BaseTableStyleHorizontalPadding).
					PaddingRight(BaseTableStyleHorizontalPadding).
					PaddingTop(1).
					Margin(1, 0)

	NameStyle           = lipgloss.NewStyle().Foreground(Light)
	ActiveStyle         = lipgloss.NewStyle().Foreground(Green)
	InactiveStyle       = lipgloss.NewStyle().Foreground(Orange)
	DefaultRowDataStyle = lipgloss.NewStyle().Foreground(Gray)
	BaseCellStyle       = lipgloss.NewRenderer(os.Stdout).NewStyle().Padding(0, 4, 1, 0)
	TableHeaderStyle    = BaseCellStyle.Copy().Foreground(LightGray).Bold(false).Padding(0).MarginRight(4)
)

var LogPrefixColors = []lipgloss.AdaptiveColor{
	Blue, Yellow, Orange, Cyan,
}

func GetStyledSelectList(items []list.Item) list.Model {

	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(Green).
		Foreground(Green).
		Bold(true).
		Padding(0, 0, 0, 1)

	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Copy().Foreground(DimmedGreen).Bold(false)

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(Green).Background(Green)
	l.Styles.Title = lipgloss.NewStyle().Foreground(Dark).Bold(true).
		Background(lipgloss.Color("#fff")).Padding(0)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(Green)

	l.SetStatusBarItemName("item\n\n"+lipgloss.NewStyle().Foreground(LightGray).Render("==="), "items\n\n"+lipgloss.NewStyle().Foreground(LightGray).Render("==="))

	return l
}

func GetCustomTheme() *huh.Theme {
	newTheme := huh.ThemeCharm()

	b := &newTheme.Blurred
	b.FocusedButton.Background(Green)
	b.FocusedButton.Bold(true)
	b.TextInput.Prompt.Foreground(Light)
	b.TextInput.Cursor.Foreground(Light)
	b.SelectSelector.Foreground(Green)
	b.Title.Foreground(Gray).Bold(true)
	b.Description.Foreground(LightGray)

	f := &newTheme.Focused
	f.Title.Foreground(Green).Bold(true)
	f.Description.Foreground(LightGray).Bold(true)
	f.FocusedButton.Bold(true)
	f.FocusedButton.Background(Green)
	f.TextInput.Prompt.Foreground(Green)
	f.TextInput.Cursor.Foreground(Light)
	f.SelectSelector.Foreground(Green)
	f.SelectedOption.Foreground(Green)

	f.Base.BorderForeground(Green)

	f.Base.MarginTop(DefaultLayoutMarginTop)
	b.Base.MarginTop(DefaultLayoutMarginTop)

	return newTheme
}

func GetInitialCommandTheme() *huh.Theme {

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
	f.Title.Foreground(Green).Bold(true)
	f.FocusedButton.Bold(true)
	f.FocusedButton.Background(Green)
	f.TextInput.Prompt.Foreground(Green)
	f.TextInput.Cursor.Foreground(Light)
	f.SelectSelector.Foreground(Green)

	f.Base.UnsetMarginLeft()
	f.Base.UnsetPaddingLeft()
	f.Base.BorderLeft(false)

	f.SelectedOption = lipgloss.NewStyle().Foreground(Green)

	return newTheme
}
