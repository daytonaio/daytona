// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mcp

import "github.com/daytonaio/daytona/cli/apiclient"

var DaytonaMCPHeaders map[string]string = map[string]string{
	apiclient.DaytonaSourceHeader: "daytona-mcp",
}
