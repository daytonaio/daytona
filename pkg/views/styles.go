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
	Red         = lipgloss.AdaptiveColor{Light: "#FF4672", Dark: "#ED567A"}
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
	TableHeaderStyle    = BaseCellStyle.Foreground(LightGray).Bold(false).Padding(0).MarginRight(4)
)

var LogPrefixColors = []lipgloss.AdaptiveColor{
	Blue, Orange, Cyan, Yellow,
}

func GetStyledSelectList(items []list.Item, parentIdentifier ...string) list.Model {

	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(Green).
		Foreground(Green).
		Bold(true).
		Padding(0, 0, 0, 1)

	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Foreground(DimmedGreen).Bold(false)

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(Green).Background(Green)
	l.Styles.Title = lipgloss.NewStyle().Foreground(Dark).Bold(true).
		Background(lipgloss.Color("#fff")).Padding(0)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(Green)

	singularItemName := "item " + lipgloss.NewStyle().Foreground(LightGray).Render("===")
	var pluralItemName string
	if len(parentIdentifier) == 0 {
		pluralItemName = "items " + lipgloss.NewStyle().Foreground(LightGray).Render("===")
	} else {
		pluralItemName = "items " + lipgloss.NewStyle().Foreground(LightGray).Render("")

		parentIdentifierStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
		formattedParentIdentifier := parentIdentifierStyle.Render("(" + parentIdentifier[0] + ")")

		pluralItemName += formattedParentIdentifier
		pluralItemName += " " + lipgloss.NewStyle().Foreground(LightGray).Render("===")
	}

	l.SetStatusBarItemName(singularItemName, pluralItemName)

	return l
}

func GetCustomTheme() *huh.Theme {
	t := huh.ThemeCharm()

	t.Blurred.FocusedButton = t.Blurred.FocusedButton.Background(Green)
	t.Blurred.FocusedButton = t.Blurred.FocusedButton.Bold(true)
	t.Blurred.TextInput.Prompt = t.Blurred.TextInput.Prompt.Foreground(Light)
	t.Blurred.TextInput.Cursor = t.Blurred.TextInput.Cursor.Foreground(Light)
	t.Blurred.SelectSelector = t.Blurred.SelectSelector.Foreground(Green)
	t.Blurred.Title = t.Blurred.Title.Foreground(Gray).Bold(true)
	t.Blurred.Description = t.Blurred.Description.Foreground(LightGray)

	t.Focused.Title = t.Focused.Title.Foreground(Green).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(LightGray).Bold(true)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Bold(true)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Background(Green)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(Green)
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(Light)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(Green)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(Green)

	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(Red)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(Red)

	t.Focused.Base = t.Focused.Base.BorderForeground(Green)
	t.Focused.Base = t.Focused.Base.BorderBottomForeground(Green)

	t.Focused.Base = t.Focused.Base.MarginTop(DefaultLayoutMarginTop)
	t.Blurred.Base = t.Blurred.Base.MarginTop(DefaultLayoutMarginTop)

	return t
}

func GetInitialCommandTheme() *huh.Theme {

	newTheme := huh.ThemeCharm()

	newTheme.Blurred.Title = newTheme.Focused.Title

	b := &newTheme.Blurred
	b.FocusedButton = b.FocusedButton.Background(Green)
	b.FocusedButton = b.FocusedButton.Bold(true)
	b.TextInput.Prompt = b.TextInput.Prompt.Foreground(Green)
	b.TextInput.Cursor = b.TextInput.Cursor.Foreground(Green)
	b.SelectSelector = b.SelectSelector.Foreground(Green)

	f := &newTheme.Focused
	f.Base = f.Base.BorderForeground(lipgloss.Color("fff"))
	f.Title = f.Title.Foreground(Green).Bold(true)
	f.FocusedButton = f.FocusedButton.Bold(true)
	f.FocusedButton = f.FocusedButton.Background(Green)
	f.TextInput.Prompt = f.TextInput.Prompt.Foreground(Green)
	f.TextInput.Cursor = f.TextInput.Cursor.Foreground(Light)
	f.SelectSelector = f.SelectSelector.Foreground(Green)

	f.Base = f.Base.UnsetMarginLeft()
	f.Base = f.Base.UnsetPaddingLeft()
	f.Base = f.Base.BorderLeft(false)

	f.SelectedOption = lipgloss.NewStyle().Foreground(Green)

	return newTheme
}
