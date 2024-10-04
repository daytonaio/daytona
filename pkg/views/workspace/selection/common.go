// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/daytonaio/daytona/pkg/views"
)

// Adds navigation options for pagination to a selection list
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

// Adds 'Load more' option to a selection list for efficient pagination
// totalItems count is exclusive of pagination options.
func AddLoadMoreOptionToList(items []list.Item, totalItems int, curPage, perPage int32) []list.Item {
	curPageItems := int32(totalItems)
	if curPage > 1 {
		curPageItems = (int32)(totalItems) - (perPage * curPage)
	}

	if curPageItems == perPage {
		items = append(items, item[string]{
			id:             views.ListNavigationText,
			title:          views.NavigationStyle.Render(views.ListNavigationRenderText),
			choiceProperty: views.ListNavigationText,
			desc:           "Loads next set of remaining items",
		})
	}

	return items
}
