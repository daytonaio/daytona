// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"testing"
)

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
			FormatFlag = tt.format
			defer func() { FormatFlag = "" }()
			if got := IsStructuredOutput(); got != tt.want {
				t.Errorf("IsStructuredOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveOutputModePicksTSVWhenNotTTY(t *testing.T) {
	// Going through resolveOutputMode would hit the real os.Stdout; instead we
	// exercise the same logic with the predicate exposed for testing.
	FormatFlag = ""
	defer func() { FormatFlag = "" }()

	applyDefaults(false /* isStdoutTTY */, "" /* NO_COLOR */)
	if FormatFlag != "tsv" {
		t.Errorf("FormatFlag = %q, want %q", FormatFlag, "tsv")
	}
}

func TestResolveOutputModeKeepsExplicitFormat(t *testing.T) {
	FormatFlag = "json"
	defer func() { FormatFlag = "" }()

	applyDefaults(false, "")
	if FormatFlag != "json" {
		t.Errorf("explicit FormatFlag should win: got %q, want %q", FormatFlag, "json")
	}
}

func TestResolveOutputModeKeepsFormatWhenTTY(t *testing.T) {
	FormatFlag = ""
	defer func() { FormatFlag = "" }()

	applyDefaults(true, "")
	if FormatFlag != "" {
		t.Errorf("interactive TTY should not auto-set format: got %q", FormatFlag)
	}
}
