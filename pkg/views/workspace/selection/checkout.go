// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type CheckoutOption struct {
	Title string
	Id    string
}

var (
	CheckoutDefault = CheckoutOption{Title: "Clone the default branch", Id: "default"}
	CheckoutBranch  = CheckoutOption{Title: "Branches", Id: "branch"}
	CheckoutPR      = CheckoutOption{Title: "Pull/Merge requests", Id: "pullrequest"}
)

func selectCheckoutPrompt(checkoutOptions []CheckoutOption, additionalProjectOrder int, choiceChan chan<- string) {
	items := []list.Item{}

	for _, checkoutOption := range checkoutOptions {
		newItem := item[string]{id: checkoutOption.Id, title: checkoutOption.Title, choiceProperty: checkoutOption.Id}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)

	title := "Cloning Options"
	if additionalProjectOrder > 1 {
		title += fmt.Sprintf(" (Project #%d)", additionalProjectOrder)
	}
	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle
	m := model[string]{list: l}

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

func GetCheckoutOptionFromPrompt(additionalProjectOrder int, checkoutOptions []CheckoutOption) CheckoutOption {
	choiceChan := make(chan string)

	go selectCheckoutPrompt(checkoutOptions, additionalProjectOrder, choiceChan)

	checkoutOptionId := <-choiceChan

	for _, checkoutOption := range checkoutOptions {
		if checkoutOption.Id == checkoutOptionId {
			return checkoutOption
		}
	}
	return CheckoutDefault
}
