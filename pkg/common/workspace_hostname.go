// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"strings"
)

func GetWorkspaceHostname(workspaceId string) string {
	// Replace special chars with hyphen to form valid hostname
	// String resulting in consecutive hyphens is also valid
	workspaceId = strings.ReplaceAll(workspaceId, "_", "-")
	workspaceId = strings.ReplaceAll(workspaceId, "*", "-")
	workspaceId = strings.ReplaceAll(workspaceId, ".", "-")

	hostname := fmt.Sprintf("ws-%s", workspaceId)

	if len(hostname) > 63 {
		return hostname[:63]
	}

	return hostname
}
