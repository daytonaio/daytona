// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package security

import "testing"

func TestValidateCommand_AllowsSimpleAllowlistedCommand(t *testing.T) {
	err := ValidateCommand("ls -la /tmp")
	if err != nil {
		t.Fatalf("expected command to be allowed, got error: %v", err)
	}
}

func TestValidateCommand_DeniesCommandChaining(t *testing.T) {
	err := ValidateCommand("ls && whoami")
	if err == nil {
		t.Fatal("expected command with chaining token to be denied")
	}
}

func TestValidateCommand_DeniesNonAllowlistedExecutable(t *testing.T) {
	err := ValidateCommand("perl -e 'print 1'")
	if err == nil {
		t.Fatal("expected non-allowlisted executable to be denied")
	}
}

func TestValidateCommand_DeniesTooLongCommand(t *testing.T) {
	command := "echo "
	for len(command) <= MaxCommandLength {
		command += "a"
	}

	err := ValidateCommand(command)
	if err == nil {
		t.Fatal("expected overly long command to be denied")
	}
}

func TestValidateCwd_DeniesControlCharacters(t *testing.T) {
	err := ValidateCwd("/tmp\n")
	if err == nil {
		t.Fatal("expected cwd with control characters to be denied")
	}
}

func TestValidateCwd_AllowsNormalPath(t *testing.T) {
	err := ValidateCwd("/workspace/project")
	if err != nil {
		t.Fatalf("expected cwd to be allowed, got error: %v", err)
	}
}
