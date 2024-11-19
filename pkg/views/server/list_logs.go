// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

func ListServerLogs(logFiles []string) {
	if len(logFiles) == 0 {
		views_util.NotifyEmptyServerLogList(true)
		return
	}

	views.RenderInfoMessageBold("Server Log Files")
	for _, logFile := range logFiles {
		views.RenderListLine("\t" + filepath.Base(logFile))
	}
}
