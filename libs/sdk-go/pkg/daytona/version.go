// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var version string

// Version is the semantic version of the Daytona SDK.
//
// This value is embedded at build time from the VERSION file.
//
// Example:
//
//	fmt.Printf("Daytona SDK version: %s\n", daytona.Version)
var Version = strings.TrimSpace(version)
