// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
)

func CheckAndAppendTimeLabel(stateLabel *string, state apiclient.ResourceState, uptime int32) {
	if state.Name == apiclient.ResourceStateNameStarted && uptime != 0 {
		*stateLabel = fmt.Sprintf("%s (%s)", *stateLabel, util.FormatUptime(uptime))
	} else if state.Name == apiclient.ResourceStateNameStopped || state.Name == apiclient.ResourceStateNameUnresponsive || state.Name == apiclient.ResourceStateNameError {
		*stateLabel = fmt.Sprintf("%s (%s)", *stateLabel, util.FormatTimestamp(state.UpdatedAt))
	}
}
