// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"strings"
)

func GetTailscaleHostname(resourceId string) string {
	// Replace special chars with hyphen to form valid hostname
	// String resulting in consecutive hyphens is also valid
	resourceId = strings.ReplaceAll(resourceId, "_", "-")
	resourceId = strings.ReplaceAll(resourceId, "*", "-")
	resourceId = strings.ReplaceAll(resourceId, ".", "-")

	if len(resourceId) > 63 {
		return resourceId[:63]
	}

	return resourceId
}
