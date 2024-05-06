// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

var CustomRepoIdentifier = "<CUSTOM_REPO>"

var selectedStyles = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder(), false, false, false, true).
	BorderForeground(views.Green).
	Bold(true).
	Padding(0, 0, 0, 1)

type item[T any] struct {
	id, title, desc, createdTime, uptime, target string
	choiceProperty                               T
}

func (i item[T]) Title() string       { return i.title }
func (i item[T]) Id() string          { return i.id }
func (i item[T]) Description() string { return i.desc }
func (i item[T]) FilterValue() string { return i.title }
func (i item[T]) CreatedTime() string { return i.createdTime }
func (i item[T]) Uptime() string      { return i.uptime }
func (i item[T]) Target() string      { return i.target }

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

		m.footer = views.GetListFooter(activeProfile.Name)
	}

	terminalWidth, terminalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return ""
	}

	return views.DocStyle.Width(terminalWidth - 4).Height(terminalHeight - 4).Render(m.list.View() + m.footer)
}

type ItemDelegate[T any] struct {
}

func (d ItemDelegate[T]) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, _ := listItem.(item[T]) // Cast the listItem to your custom item type
	s := strings.Builder{}

	var isSelected = index == m.Index()

	baseStyles := lipgloss.NewStyle().Padding(0, 0, 0, 2)

	title := baseStyles.Copy().Render(i.Title())
	idWithTargetString := fmt.Sprintf("%s (%s)", i.Id(), i.Target())
	idWithTarget := baseStyles.Copy().Foreground(views.Gray).Render(idWithTargetString)
	description := baseStyles.Copy().Render(i.Description())

	// Add the created/updated time if it's available
	timeWidth := m.Width() - baseStyles.GetHorizontalFrameSize() - lipgloss.Width(title)
	timeStyles := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Width(timeWidth)
	timeString := timeStyles.Render("")
	if i.Uptime() != "" {
		timeString = timeStyles.Render(i.Uptime())
	} else if i.CreatedTime() != "" {
		timeString = timeStyles.Render(fmt.Sprintf("created %s", i.CreatedTime()))
	}

	// Adjust styles as the user moves through the menu
	if isSelected {
		title = selectedStyles.Copy().Foreground(views.Green).Render(i.Title())
		idWithTarget = selectedStyles.Copy().Foreground(views.Gray).Render(idWithTargetString)
		description = selectedStyles.Copy().Foreground(views.DimmedGreen).Render(i.Description())
		timeString = timeStyles.Copy().Foreground(views.DimmedGreen).Render(timeString)
	}

	// Render to the terminal
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Bottom, title, timeString))
	s.WriteRune('\n')
	s.WriteString(idWithTarget)
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
	return nil
}
