// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"strings"
)

func ValidateImageName(imageName string) error {
	parts := strings.Split(imageName, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid image format: must contain exactly one colon (e.g., 'ubuntu:22.04')")
	}
	if parts[1] == "latest" {
		return fmt.Errorf("tag 'latest' not allowed, please use a specific version tag")
	}

	return nil
}
