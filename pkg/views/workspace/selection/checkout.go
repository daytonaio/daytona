// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func selectCheckoutPrompt(checkoutOptions []gitprovider.CheckoutOption, secondaryProjectOrder int, choiceChan chan<- string) {
	items := []list.Item{}

	for _, checkoutOption := range checkoutOptions {
		newItem := item[string]{id: checkoutOption.Id, title: checkoutOption.Title, choiceProperty: checkoutOption.Id}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)
	m := model[string]{list: l}
	m.list.Title = "CLONING OPTIONS"
	if secondaryProjectOrder > 0 {
		m.list.Title += fmt.Sprintf(" (Secondary Project #%d)", secondaryProjectOrder)
	}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[string]); ok && m.choice != nil {
		choiceChan <- *m.choice
	} else {
		choiceChan <- ""
	}
}

func GetCheckoutOptionFromPrompt(secondaryProjectOrder int, checkoutOptions []gitprovider.CheckoutOption) gitprovider.CheckoutOption {
	choiceChan := make(chan string)

	go selectCheckoutPrompt(checkoutOptions, secondaryProjectOrder, choiceChan)

	checkoutOptionId := <-choiceChan

	for _, checkoutOption := range checkoutOptions {
		if checkoutOption.Id == checkoutOptionId {
			return checkoutOption
		}
	}
	return gitprovider.CheckoutDefault
}
