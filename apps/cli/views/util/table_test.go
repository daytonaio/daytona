// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"strings"
	"testing"

	"github.com/daytonaio/daytona/cli/internal"
)

func TestGetTableViewTSVStripsANSIAndUsesTabs(t *testing.T) {
	internal.FormatFlag = "tsv"
	defer func() { internal.FormatFlag = "" }()

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
	internal.FormatFlag = "tsv"
	defer func() { internal.FormatFlag = "" }()

	out := GetTableView(nil, []string{"Sandbox"}, nil, func() {})
	if out != "" {
		t.Errorf("empty data should yield empty string, got %q", out)
	}

	out = GetTableView([][]string{}, []string{"Sandbox"}, nil, func() {})
	if out != "" {
		t.Errorf("zero-length data should yield empty string, got %q", out)
	}
}

// TestGetTableViewTSVResistsRowInjection pins that a single input row
// containing literal tab/newline/CR characters in any cell still yields
// exactly one output line. Without sanitization an attacker who controls
// a sandbox name field could inject fake rows that downstream `awk` /
// `xargs` pipelines would act on.
func TestGetTableViewTSVResistsRowInjection(t *testing.T) {
	internal.FormatFlag = "tsv"
	defer func() { internal.FormatFlag = "" }()

	rows := [][]string{
		{"sb-victim\tSTARTED\tus\tlarge\tINJECTED\nsb-attacker", "real-state", "eu"},
		{"normal\rrow", "ok", "us"},
	}
	out := GetTableView(rows, []string{"a", "b", "c"}, nil, func() {})

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d output lines, want 2 (one per input row); output=%q", len(lines), out)
	}
	for i, line := range lines {
		if strings.ContainsAny(line, "\n\r") {
			t.Errorf("line %d contains stray CR/LF after sanitization: %q", i, line)
		}
		if got := strings.Count(line, "\t"); got != 2 {
			t.Errorf("line %d has %d tabs, want 2 (3 columns): %q", i, got, line)
		}
	}
}

// TestStripANSICoversOSCAndDCS pins that StripANSI catches the full ANSI
// family — not just CSI color codes. OSC-8 hyperlinks in user-controlled
// strings are a phishing vector if they survive the strip.
func TestStripANSICoversOSCAndDCS(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"OSC-8 hyperlink", "\x1b]8;;https://evil.example/\x07click\x1b]8;;\x07", "click"},
		{"OSC-0 title set", "\x1b]0;malicious-title\x07after", "after"},
		{"OSC-52 clipboard (ST-terminated)", "\x1b]52;c;abc\x1b\\after", "after"},
		{"DCS sequence", "\x1bP1;2$rresponse\x1b\\after", "after"},
		{"2-byte ESC (charset designation)", "\x1b(Bnormal", "normal"},
		{"CSI cursor (legacy regression)", "\x1b[2K\x1b[Hcleared", "cleared"},
		{"stray bare ESC", "before\x1bafter", "beforeafter"},
		{"plain text unchanged", "hello world", "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripANSI(tt.in); got != tt.want {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestSanitizeTSVStripsControlChars verifies the row-/field-injection
// defense composed on top of StripANSI.
func TestSanitizeTSVStripsControlChars(t *testing.T) {
	in := "a\tb\nc\rd\x1b[31me\x1b[0m\x1b]8;;u\x07f\x1b]8;;\x07"
	got := SanitizeTSV(in)

	if strings.ContainsAny(got, "\t\n\r\x1b") {
		t.Errorf("SanitizeTSV result contains delimiter or escape bytes: %q", got)
	}
	for _, want := range []string{"a", "b", "c", "d", "e", "f"} {
		if !strings.Contains(got, want) {
			t.Errorf("SanitizeTSV dropped legitimate content %q from %q (got %q)", want, in, got)
		}
	}
}
