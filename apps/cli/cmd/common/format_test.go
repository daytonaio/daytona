// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

// withColorProfile saves the current default lipgloss color profile and
// restores it after the test, so applyDefaults doesn't leak global state.
func withColorProfile(t *testing.T) {
	t.Helper()
	prev := lipgloss.DefaultRenderer().ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
}

func TestIsStructuredOutput(t *testing.T) {
	tests := []struct {
		format string
		want   bool
	}{
		{"", false},
		{"tsv", false},
		{"json", true},
		{"yaml", true},
		{"unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			internal.FormatFlag = tt.format
			defer func() { internal.FormatFlag = "" }()
			if got := internal.IsStructuredOutput(); got != tt.want {
				t.Errorf("internal.IsStructuredOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveOutputModePicksTSVWhenNotTTY(t *testing.T) {
	// Going through resolveOutputMode would hit the real os.Stdout; instead we
	// exercise the same logic with the predicate exposed for testing.
	internal.FormatFlag = ""
	defer func() { internal.FormatFlag = "" }()

	applyDefaults(false /* isStdoutTTY */, "" /* NO_COLOR */)
	if internal.FormatFlag != "tsv" {
		t.Errorf("internal.FormatFlag = %q, want %q", internal.FormatFlag, "tsv")
	}
}

func TestResolveOutputModeKeepsExplicitFormat(t *testing.T) {
	internal.FormatFlag = "json"
	defer func() { internal.FormatFlag = "" }()

	applyDefaults(false, "")
	if internal.FormatFlag != "json" {
		t.Errorf("explicit internal.FormatFlag should win: got %q, want %q", internal.FormatFlag, "json")
	}
}

func TestResolveOutputModeKeepsFormatWhenTTY(t *testing.T) {
	internal.FormatFlag = ""
	defer func() { internal.FormatFlag = "" }()

	applyDefaults(true, "")
	if internal.FormatFlag != "" {
		t.Errorf("interactive TTY should not auto-set format: got %q", internal.FormatFlag)
	}
}

// TestApplyDefaultsNoColorStripsInTTY pins the headline acceptance criterion:
// NO_COLOR=1 must strip ANSI even when stdout is an interactive terminal.
// We assert the lipgloss color profile directly (not just internal.FormatFlag).
func TestApplyDefaultsNoColorStripsInTTY(t *testing.T) {
	withColorProfile(t)
	internal.FormatFlag = ""
	defer func() { internal.FormatFlag = "" }()

	// Seed a non-Ascii profile so we can verify the call actually changes it.
	lipgloss.SetColorProfile(termenv.TrueColor)

	applyDefaults(true /* isStdoutTTY */, "1" /* NO_COLOR */)

	if got := lipgloss.DefaultRenderer().ColorProfile(); got != termenv.Ascii {
		t.Errorf("NO_COLOR in TTY should set profile to Ascii, got %v", got)
	}
	if internal.FormatFlag != "" {
		t.Errorf("NO_COLOR alone should not auto-set format: got %q", internal.FormatFlag)
	}
}

// TestApplyDefaultsPipedSetsAsciiProfile pins that piped stdout also strips
// colors (not just sets tsv). Same headline guarantee, different path.
func TestApplyDefaultsPipedSetsAsciiProfile(t *testing.T) {
	withColorProfile(t)
	internal.FormatFlag = ""
	defer func() { internal.FormatFlag = "" }()

	lipgloss.SetColorProfile(termenv.TrueColor)

	applyDefaults(false, "")

	if got := lipgloss.DefaultRenderer().ColorProfile(); got != termenv.Ascii {
		t.Errorf("piped stdout should set profile to Ascii, got %v", got)
	}
}

// TestRegisterFormatFlagPreRunDoesNotBlockStdoutForTSV pins a regression-class
// concern: BlockStdOut() sets os.Stdout = nil, which silently discards all
// writes. Before this change, any non-empty FormatFlag triggered it, which
// would have swallowed our new TSV output. The gate is now IsStructuredOutput()
// so tsv falls through with stdout intact.
func TestRegisterFormatFlagPreRunDoesNotBlockStdoutForTSV(t *testing.T) {
	origStdout := os.Stdout
	t.Cleanup(func() {
		os.Stdout = origStdout
		standardOut = nil
		internal.FormatFlag = ""
	})

	cmd := &cobra.Command{Use: "fake"}
	RegisterFormatFlag(cmd)

	internal.FormatFlag = "tsv"
	cmd.PreRun(cmd, nil)

	if os.Stdout == nil {
		t.Fatal("PreRun should not BlockStdOut for tsv; got nil os.Stdout")
	}
}

// TestRegisterFormatFlagPreRunBlocksStdoutForJSON pins the *other* side of
// the gating: json/yaml must still trigger BlockStdOut so version-mismatch
// warnings (and other stray writes) don't contaminate structured output.
func TestRegisterFormatFlagPreRunBlocksStdoutForJSON(t *testing.T) {
	origStdout := os.Stdout
	t.Cleanup(func() {
		os.Stdout = origStdout
		standardOut = nil
		internal.FormatFlag = ""
	})

	cmd := &cobra.Command{Use: "fake"}
	RegisterFormatFlag(cmd)

	internal.FormatFlag = "json"
	cmd.PreRun(cmd, nil)

	if os.Stdout != nil {
		t.Fatal("PreRun should BlockStdOut for json; os.Stdout was not nilled")
	}
}
