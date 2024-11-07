// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/daytonaio/daytona/pkg/views"
)

// Adds 'Load more' option to a selection list for efficient pagination
// totalItems count is exclusive of pagination options.
func AddLoadMoreOptionToList(items []list.Item) []list.Item {
	items = append(items, item[string]{
		id:             views.ListNavigationText,
		title:          views.ListNavigationRenderText,
		choiceProperty: views.ListNavigationText,
		desc:           "Loads next set of remaining items",
	})

	return items
}
