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
	"github.com/daytonaio/daytona/pkg/apiclient"
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
	id, title, desc, targetName, repository, createdTime, state string
	workspace                                                   *apiclient.WorkspaceDTO
	choiceProperty                                              T
	isMarked                                                    bool
	isMultipleSelect                                            bool
	isDisabled                                                  bool
	action                                                      string
}

func (i item[T]) Title() string       { return i.title }
func (i item[T]) Id() string          { return i.id }
func (i item[T]) Description() string { return i.desc }
func (i item[T]) TargetName() string  { return i.targetName }
func (i item[T]) Repository() string  { return i.repository }
func (i item[T]) CreatedTime() string { return i.createdTime }
func (i item[T]) State() string       { return i.state }
func (i item[T]) FilterValue() string { return i.title }

type model[T any] struct {
	list            list.Model
	choice          *T
	choices         []*T
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
		case "down", "j":
			for {
				isSubsequentItemsDisabled := true
				curIndex := m.list.Index()
				lastIndex := len(m.list.Items()) - 1
				for i := curIndex + 1; i <= lastIndex; i++ {
					if !m.list.Items()[i].(item[T]).isDisabled {
						isSubsequentItemsDisabled = false
						break
					}
				}
				// no need of moving cursor down if all subsequent items after current index have
				// disabled property true.
				if isSubsequentItemsDisabled {
					break
				}
				m.list.CursorDown()
				item := m.list.SelectedItem().(item[T])
				if !item.isDisabled {
					break
				}
			}
			return m, nil

		case "up", "k":
			for {
				isPrecedingItemsDisabled := true
				curIndex := m.list.Index()
				for i := curIndex - 1; i >= 0; i-- {
					if !m.list.Items()[i].(item[T]).isDisabled {
						isPrecedingItemsDisabled = false
						break
					}
				}
				// no need of moving cursor up if all preceding items before current index have
				// disabled property true.
				if isPrecedingItemsDisabled {
					break
				}
				m.list.CursorUp()
				item := m.list.SelectedItem().(item[T])
				if !item.isDisabled {
					break
				}
			}
			return m, nil

		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item[T])
			if ok && !i.isDisabled {
				m.choice = &i.choiceProperty
			}
			workspaceList := m.list.Items()
			var choices []*T
			for _, workspace := range workspaceList {
				if workspace.(item[T]).isMarked {
					workspaceItem, ok := workspace.(item[T])
					if !ok || workspaceItem.isDisabled {
						continue
					}
					choices = append(choices, &workspaceItem.choiceProperty)
				}

			}
			m.choices = choices
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
	idWithTargetString := fmt.Sprintf("%s (%s)", i.Id(), i.TargetName())
	if i.Id() == NewWorkspaceIdentifier {
		idWithTargetString = ""
	}
	idWithTarget := baseStyles.Foreground(views.Gray).Render(idWithTargetString)
	repository := baseStyles.Foreground(views.DimmedGreen).Render(i.Repository())
	description := baseStyles.Render(i.Description())

	// Add the created/updated time if it's available
	timeWidth := m.Width() - baseStyles.GetHorizontalFrameSize() - lipgloss.Width(title)
	timeStyles := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Width(timeWidth)
	stateLabel := timeStyles.Render("")
	if i.State() != "" {
		stateLabel = timeStyles.Render(i.State())
	}

	if i.isDisabled {
		title = baseStyles.Foreground(views.Gray).Render(i.Title())
		idWithTarget = baseStyles.Foreground(views.Gray).Render(idWithTargetString)
		repository = baseStyles.Foreground(views.Gray).Render(i.Repository())
		description = baseStyles.Foreground(views.Gray).Render(i.Description())
		stateLabel = timeStyles.Foreground(views.Gray).Render(stateLabel)
	}

	// Adjust styles as the user moves through the menu
	if isSelected {
		if !i.isDisabled {
			title = selectedStyles.Foreground(views.Green).Render(i.Title())
			idWithTarget = selectedStyles.Foreground(views.Gray).Render(idWithTargetString)
			repository = selectedStyles.Foreground(views.DimmedGreen).Render(i.Repository())
			description = selectedStyles.Foreground(views.DimmedGreen).Render(i.Description())
			stateLabel = timeStyles.Foreground(views.DimmedGreen).Render(stateLabel)
		} else {
			title = selectedStyles.Render(i.Title())
			idWithTarget = selectedStyles.Foreground(views.Gray).Render(idWithTargetString)
			repository = selectedStyles.Foreground(views.LightGray).Render(i.Repository())
			description = selectedStyles.Foreground(views.LightGray).Render(i.Description())
			stateLabel = timeStyles.Foreground(views.LightGray).Render(stateLabel)
		}
	}

	// Render to the terminal
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Bottom, title, stateLabel))
	s.WriteRune('\n')
	s.WriteString(idWithTarget)
	s.WriteRune('\n')
	s.WriteString(repository)
	s.WriteRune('\n')
	if i.Description() != "" {
		s.WriteString(description)
		s.WriteRune('\n')
	}

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
	i, ok := m.SelectedItem().(item[T])
	if !ok {
		return nil
	}

	m.StatusMessageLifetime = time.Millisecond * 2000

	var title string
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "x":
			if !i.isMultipleSelect {
				return nil
			}
			if i.isDisabled {
				return m.NewStatusMessage(statusMessageGreenStyle("Workspace ") + i.title + statusMessageGreenStyle(" selection is disabled for this action"))
			}
			if i.isMarked {
				i.title = strings.TrimPrefix(i.title, statusMessageDangerStyle(fmt.Sprintf("%s: ", i.action)))
				i.isMarked = false
				m.SetItem(m.Index(), i)
				return m.NewStatusMessage(statusMessageGreenStyle("Removed workspace from list: ") + statusMessageGreenStyle(i.title))
			}

			title = i.title
			i.title = statusMessageDangerStyle(fmt.Sprintf("%s: ", i.action)) + statusMessageGreenStyle(i.title)
			i.isMarked = true
			m.SetItem(m.Index(), i)
			return m.NewStatusMessage(statusMessageDangerStyle("Added workspace to list: ") + statusMessageGreenStyle(title))
		}
	}
	return nil
}
