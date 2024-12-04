// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

func GetStateSortPriorities(state1, state2 apiclient.ModelsResourceStateName) (int, int) {
	pi, ok := views.ResourceListStatePriorities[state1]
	if !ok {
		pi = 99
	}
	pj, ok2 := views.ResourceListStatePriorities[state2]
	if !ok2 {
		pj = 99
	}

	return pi, pj
}
