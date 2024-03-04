// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

const NewTargetName = "+ New Target"

func GetTargetFromPrompt(targets []serverapiclient.TargetDTO, withNewTarget bool) (*serverapiclient.TargetDTO, error) {
	items := util.ArrayMap(targets, func(t serverapiclient.TargetDTO) list.Item {
		return item{
			target: t,
		}
	})

	if withNewTarget {
		name := NewTargetName
		options := "{}"
		items = append(items, item{
			target: serverapiclient.TargetDTO{
				Name:    &name,
				Options: &options,
			},
		})
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = "Choose a target"

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model); ok && m.choice != nil {
		return m.choice, nil
	}

	return nil, errors.New("no target selected")
}
