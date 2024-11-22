// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

var selectedStyles = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder(), false, false, false, true).
	BorderForeground(views.Green).
	Bold(true).
	Padding(0, 0, 0, 1)

var statusMessageGreenStyle = lipgloss.NewStyle().Bold(true).
	Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
	Render

var statusMessageDangerStyle = lipgloss.NewStyle().Bold(true).
	Foreground(lipgloss.AdaptiveColor{Light: "#FF474C", Dark: "#FF474C"}).
	Render

type item[T any] struct {
	title, desc    string
	choiceProperty T
}

func (i item[T]) Title() string       { return i.title }
func (i item[T]) Description() string { return i.desc }
func (i item[T]) FilterValue() string { return i.title }

type model[T any] struct {
	list            list.Model
	choice          *T
	footer          string
	initialWidthSet bool
}

func (m model[T]) Init() tea.Cmd {
	return nil
}

func (m model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.initialWidthSet {
		_, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			m.list.SetWidth(150)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item[T])
			if ok {
				m.choice = &i.choiceProperty
			}

			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		h, v := views.DocStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model[T]) View() string {
	if m.footer == "" {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		m.footer = views.GetListFooter(activeProfile.Name, views.DefaultListFooterPadding)
	}

	terminalWidth, terminalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return ""
	}

	if m.list.FilterState() == list.Filtering {
		return views.DocStyle.MaxWidth(terminalWidth - 4).MaxHeight(terminalHeight - 4).Render(m.list.View() + m.footer)
	}

	return views.DocStyle.MaxWidth(terminalWidth - 4).Height(terminalHeight - 2).Render(m.list.View() + m.footer)
}

type ItemDelegate[T any] struct {
}

func (d ItemDelegate[T]) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, _ := listItem.(item[T]) // Cast the listItem to your custom item type
	s := strings.Builder{}

	var isSelected = index == m.Index()

	baseStyles := lipgloss.NewStyle().Padding(0, 0, 0, 2)

	title := baseStyles.Render(i.Title())
	description := baseStyles.Render(i.Description())

	// Adjust styles as the user moves through the menu
	if isSelected {
		title = selectedStyles.Foreground(views.Green).Render(i.Title())
		description = selectedStyles.Foreground(views.DimmedGreen).Render(i.Description())
	}

	// Render to the terminal
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Bottom, title))
	s.WriteRune('\n')
	s.WriteString(description)
	s.WriteRune('\n')

	fmt.Fprint(w, s.String())
}

func (d ItemDelegate[T]) Height() int {
	height := lipgloss.NewStyle().GetVerticalFrameSize() + 4
	return height
}

func (d ItemDelegate[T]) Spacing() int {
	return 0
}

func (d ItemDelegate[T]) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	_, ok := m.SelectedItem().(item[T])
	if !ok {
		return nil
	}

	m.StatusMessageLifetime = time.Millisecond * 2000

	return nil
}
