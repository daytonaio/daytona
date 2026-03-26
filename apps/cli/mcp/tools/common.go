// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import "github.com/daytonaio/daytona/cli/apiclient"

var daytonaMCPHeaders map[string]string = map[string]string{
	apiclient.DaytonaSourceHeader: "daytona-mcp",
}
