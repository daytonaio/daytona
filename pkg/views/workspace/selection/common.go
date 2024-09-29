// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/daytonaio/daytona/pkg/views"
)

// Add navigation options for pagination to a selection list
// totalItems count is exclusive of pagination options.
func AddNavigationOptionsToList(items []list.Item, totalItems int, curPage, perPage int32) []list.Item {
	if curPage > 1 {
		items = append([]list.Item{item[string]{
			id:             "prev",
			title:          views.NavigationStyle.Render("Previous Page"),
			choiceProperty: "prev",
			desc:           "Go to the previous page",
		}}, items...)
	}

	if totalItems == int(perPage) {
		items = append(items, item[string]{
			id:             "next",
			title:          views.NavigationStyle.Render("Next Page"),
			choiceProperty: "next",
			desc:           "Go to the next page",
		})
	}

	return items
}
