// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
)

func GetLogFileFromPrompt(files []string) *string {
	choiceChan := make(chan *string)

	go selectLogFileFromPrompt(files, choiceChan)

	return <-choiceChan
}

func selectLogFileFromPrompt(files []string, choiceChan chan<- *string) {
	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	items := []list.Item{}

	for _, logFile := range files {
		newItem := item[string]{title: filepath.Base(logFile), desc: "", choiceProperty: logFile}
		items = append(items, newItem)
	}

	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(views.Green).
		Foreground(views.Green).
		Bold(true).
		Padding(0, 0, 0, 1)

	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Foreground(views.DimmedGreen)

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	m := model[string]{list: l}

	m.list.Title = views.GetStyledMainTitle("Select a Log File To Read From")
	m.list.Styles.Title = lipgloss.NewStyle().Foreground(views.Green).Bold(true)

	m.footer = lipgloss.NewStyle().Bold(true).PaddingLeft(2).Render("\n\nPress 'enter' to select currentlog file.")

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[string]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}
