// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var ModelInstance model

var (
	titleStyle      = lipgloss.NewStyle().Foreground(views.Green).Bold(true)
	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle   = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	docStyle        = lipgloss.NewStyle().Margin(1, 2)
)

type item struct {
	id, name string
}

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int {
	return lipgloss.NewStyle().GetVerticalFrameSize() + 2
}
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, _ := listItem.(item)
	s := strings.Builder{}

	var isSelected = index == m.Index()
	itemStyles := lipgloss.NewStyle().Padding(0, 0, 0, 2)

	ideString := itemStyles.Copy().Render(i.name)

	if isSelected {
		selectedItemStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(views.Blue).
			Bold(true).
			Padding(0, 0, 0, 1)
		ideString = selectedItemStyle.Copy().Foreground(views.Blue).Render(i.name)
	}
	s.WriteString(ideString)
	s.WriteRune('\n')
	fmt.Fprint(w, s.String())
}

type model struct {
	list     list.Model
	choice   item
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice.id != "" {
		return ""
	}
	if m.quitting {
		return quitTextStyle.Render("Canceled")
	}
	return docStyle.Render(m.list.View())
}

func Render(ideList []config.Ide, choiceChan chan<- string) {
	items := util.ArrayMap(ideList, func(ide config.Ide) list.Item {
		return item{id: ide.Id, name: ide.Name}
	})

	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = lipgloss.NewStyle().Foreground(views.Green).Bold(true).Render("Choose your default IDE")
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	ModelInstance = model{list: l}

	m, err := tea.NewProgram(ModelInstance, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := m.(model); ok && m.choice.id != "" {
		choiceChan <- m.choice.id
	} else {
		choiceChan <- ""
	}
}
