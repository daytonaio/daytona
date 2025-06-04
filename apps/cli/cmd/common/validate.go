// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"strings"
)

func ValidateSnapshotName(snapshotName string) error {
	parts := strings.Split(snapshotName, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid snapshot format: must contain exactly one colon (e.g., 'mysnapshot:1.0')")
	}
	if parts[1] == "latest" {
		return fmt.Errorf("tag 'latest' not allowed, please use a specific version tag")
	}

	return nil
}
