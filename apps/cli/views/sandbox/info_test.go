// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"bytes"
	"strings"
	"testing"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestRenderTSVInfoEmitsCuratedKeys(t *testing.T) {
	state := apiclient.SANDBOXSTATE_STARTED
	snapshot := "node:20"
	createdAt := "2026-05-01T10:00:00Z"
	lastActivity := "2026-05-23T15:30:00Z"
	class := "small"

	sb := &apiclient.Sandbox{
		Id:             "sb-abc",
		State:          &state,
		Snapshot:       &snapshot,
		Target:         "us",
		Class:          &class,
		CreatedAt:      &createdAt,
		LastActivityAt: &lastActivity,
		Labels:         map[string]string{"env": "prod"},
	}

	var buf bytes.Buffer
	renderTSVInfo(&buf, sb)
	out := buf.String()

	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if !strings.Contains(line, "\t") {
			t.Errorf("every line should contain a tab; got %q", line)
		}
	}

	wantPairs := map[string]string{
		"id":         "sb-abc",
		"state":      "started",
		"snapshot":   "node:20",
		"region":     "us",
		"class":      "small",
		"created":    "2026-05-01T10:00:00Z",
		"last_event": "2026-05-23T15:30:00Z",
		"label.env":  "prod",
	}
	got := parseTSVPairs(out)
	for k, want := range wantPairs {
		if got[k] != want {
			t.Errorf("key %q = %q, want %q (full output:\n%s)", k, got[k], want, out)
		}
	}
}

// TestRenderTSVInfoOmitsNilFields verifies optional fields drop out of the
// output entirely when nil (no stray "snapshot\t" line, etc.).
func TestRenderTSVInfoOmitsNilFields(t *testing.T) {
	sb := &apiclient.Sandbox{Id: "sb-min", Target: "eu"}

	var buf bytes.Buffer
	renderTSVInfo(&buf, sb)
	out := buf.String()

	for _, banned := range []string{"state\t", "snapshot\t", "class\t", "created\t", "last_event\t", "label."} {
		if strings.Contains(out, banned) {
			t.Errorf("output should not contain %q for a minimal sandbox; got:\n%s", banned, out)
		}
	}
}

// TestRenderTSVInfoSanitizesAdversarialFields pins that user-controlled
// fields containing tab/newline/ANSI bytes cannot inject extra rows or
// terminal-injection sequences into the TSV output. Sandbox names and
// labels are the most likely attacker-controlled inputs in practice.
func TestRenderTSVInfoSanitizesAdversarialFields(t *testing.T) {
	sb := &apiclient.Sandbox{
		Id:     "sb\tinject\nfake",
		Target: "us\x1b]8;;https://evil/\x07click\x1b]8;;\x07",
		Labels: map[string]string{
			"key\twith\ttab": "val\nwith\nnewline",
			"normal":         "ok",
		},
	}

	var buf bytes.Buffer
	renderTSVInfo(&buf, sb)
	out := buf.String()

	if strings.ContainsRune(out, '\x1b') {
		t.Errorf("output contains stray ESC byte after sanitization: %q", out)
	}

	// The structural invariant we care about: every emitted line is exactly
	// one `key<TAB>value` pair. If sanitization had failed, the malicious id
	// or label value would have introduced extra newlines and extra tabs,
	// producing rows with !=1 tab.
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	for i, line := range lines {
		if got := strings.Count(line, "\t"); got != 1 {
			t.Errorf("line %d should have exactly 1 tab (key<TAB>value); got %d in %q", i, got, line)
		}
	}

	// Sandbox has 1 id + 1 region + 2 labels = 4 lines (no state/snapshot/etc set).
	// Without sanitization, the adversarial id+labels would expand to ~10+ rows.
	if len(lines) != 4 {
		t.Errorf("expected exactly 4 lines (id, region, 2 labels); got %d:\n%s", len(lines), out)
	}
}

// TestRenderTSVInfoLastEventFallback exercises the LastActivityAt/UpdatedAt
// fallback branch (LastActivityAt nil → use UpdatedAt).
func TestRenderTSVInfoLastEventFallback(t *testing.T) {
	updated := "2026-05-23T15:30:00Z"
	sb := &apiclient.Sandbox{Id: "sb-fb", Target: "us", UpdatedAt: &updated}

	var buf bytes.Buffer
	renderTSVInfo(&buf, sb)

	pairs := parseTSVPairs(buf.String())
	if pairs["last_event"] != updated {
		t.Errorf("last_event = %q, want %q", pairs["last_event"], updated)
	}
}

// parseTSVPairs splits TSV output into a key→value map for assertion ergonomics.
// Behavior on duplicate keys (e.g. multiple labels): last write wins. Tests
// that need to assert on multiple labels should walk the output directly.
func parseTSVPairs(out string) map[string]string {
	pairs := map[string]string{}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		k, v, _ := strings.Cut(line, "\t")
		pairs[k] = v
	}
	return pairs
}
