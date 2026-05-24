// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"strings"
	"testing"

	commoncmd "github.com/daytonaio/daytona/cli/cmd/common"
)

func TestGetTableViewTSVStripsANSIAndUsesTabs(t *testing.T) {
	commoncmd.FormatFlag = "tsv"
	defer func() { commoncmd.FormatFlag = "" }()

	rows := [][]string{
		{"\x1b[1;32msb-abc\x1b[0m", "STARTED", "us"},
		{"sb-def", "STOPPED", "eu"},
	}

	out := GetTableView(rows, []string{"Sandbox", "State", "Region"}, nil, func() {})

	if strings.ContainsRune(out, '\x1b') {
		t.Errorf("TSV output contains ANSI escape bytes: %q", out)
	}
	if !strings.Contains(out, "sb-abc") {
		t.Errorf("TSV output missing first row id: %q", out)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d rows, want 2; output=%q", len(lines), out)
	}
	for i, line := range lines {
		if got := strings.Count(line, "\t"); got != 2 {
			t.Errorf("row %d has %d tabs, want 2 (3 columns): %q", i, got, line)
		}
	}
}

func TestStripANSI(t *testing.T) {
	in := "\x1b[1;32mhello\x1b[0m world"
	want := "hello world"
	if got := StripANSI(in); got != want {
		t.Errorf("StripANSI(%q) = %q, want %q", in, got, want)
	}
}

// TestGetTableViewTSVEmptyData pins the acceptance criterion:
// "Empty list in piped mode: zero bytes on stdout, exit code 0."
func TestGetTableViewTSVEmptyData(t *testing.T) {
	commoncmd.FormatFlag = "tsv"
	defer func() { commoncmd.FormatFlag = "" }()

	out := GetTableView(nil, []string{"Sandbox"}, nil, func() {})
	if out != "" {
		t.Errorf("empty data should yield empty string, got %q", out)
	}

	out = GetTableView([][]string{}, []string{"Sandbox"}, nil, func() {})
	if out != "" {
		t.Errorf("zero-length data should yield empty string, got %q", out)
	}
}
