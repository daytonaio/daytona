// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mcp

import "testing"

func TestConfigCmdUse(t *testing.T) {
	if ConfigCmd.Use != "config" {
		t.Errorf("ConfigCmd.Use = %q, want %q", ConfigCmd.Use, "config")
	}
}
