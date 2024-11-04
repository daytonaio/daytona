// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func selectSamplePrompt(samples []apiclient.Sample, choiceChan chan<- *apiclient.Sample) {
	items := []list.Item{}

	for _, sample := range samples {
		newItem := item[apiclient.Sample]{id: sample.Name, title: sample.Name, desc: sample.GitUrl, choiceProperty: sample}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)

	title := "Choose a Sample"
	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle
	m := model[apiclient.Sample]{list: l}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[apiclient.Sample]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}

func GetSampleFromPrompt(samples []apiclient.Sample) *apiclient.Sample {
	choiceChan := make(chan *apiclient.Sample)

	go selectSamplePrompt(samples, choiceChan)

	return <-choiceChan
}
