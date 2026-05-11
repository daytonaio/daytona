// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestEmitLimitsInSync guards against drift between the JS host, the Python
// worker, and the Go reader cap. The interpreters' emit-side caps
// (MAX_CHUNK_BYTES / MAX_LINE_BYTES) are duplicated across two source files in
// two languages; this test reads them straight out of the source so a change in
// one place that isn't mirrored in the other fails at build/test time rather
// than silently letting an oversized line reach the Go reader's drain-and-skip
// recovery (which the airtight emit guard relies on staying unreachable).
func TestEmitLimitsInSync(t *testing.T) {
	jsChunk := extractShiftConst(t, "repl_host.js", "MAX_CHUNK_BYTES")
	jsLine := extractShiftConst(t, "repl_host.js", "MAX_LINE_BYTES")
	pyChunk := extractShiftConst(t, "repl_worker.py", "MAX_CHUNK_BYTES")
	pyLine := extractShiftConst(t, "repl_worker.py", "MAX_LINE_BYTES")

	if jsChunk != pyChunk {
		t.Errorf("MAX_CHUNK_BYTES mismatch: repl_host.js=%d repl_worker.py=%d", jsChunk, pyChunk)
	}
	if jsLine != pyLine {
		t.Errorf("MAX_LINE_BYTES mismatch: repl_host.js=%d repl_worker.py=%d", jsLine, pyLine)
	}

	// Both emit-side caps MUST stay at or below the Go reader's cap so a
	// well-behaved frame can never trip the reader's oversized-line recovery.
	for _, c := range []struct {
		name string
		val  int64
	}{
		{"repl_host.js MAX_CHUNK_BYTES", jsChunk},
		{"repl_host.js MAX_LINE_BYTES", jsLine},
		{"repl_worker.py MAX_CHUNK_BYTES", pyChunk},
		{"repl_worker.py MAX_LINE_BYTES", pyLine},
	} {
		if c.val > maxWorkerLineBytes {
			t.Errorf("%s=%d exceeds maxWorkerLineBytes=%d", c.name, c.val, maxWorkerLineBytes)
		}
	}

	// Sanity: the per-frame line cap should never be smaller than a single
	// chunk cap, or a lone in-budget chunk could trip the whole-line guard.
	if jsLine < jsChunk {
		t.Errorf("MAX_LINE_BYTES (%d) < MAX_CHUNK_BYTES (%d)", jsLine, jsChunk)
	}
}

// extractShiftConst reads the named source file (relative to this test's
// package dir) and extracts the value assigned to `name`. The literal may be a
// plain integer (e.g. 1048576) or a shift expression (e.g. `1 << 20`); both
// JS (`const NAME = 1 << 20`) and Python (`NAME = 1 << 20`) assignment forms
// are tolerated.
func extractShiftConst(t *testing.T, file, name string) int64 {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	// Match e.g. `MAX_LINE_BYTES = 4 << 20` (optionally `const`), capturing the
	// right-hand literal up to a comment or end of line.
	re := regexp.MustCompile(`(?m)^\s*(?:const\s+)?` + regexp.QuoteMeta(name) + `\s*=\s*([0-9]+(?:\s*<<\s*[0-9]+)?)`)
	m := re.FindSubmatch(data)
	if m == nil {
		t.Fatalf("could not find %s in %s", name, file)
	}
	return evalShift(t, file, name, string(m[1]))
}

// evalShift evaluates a literal of the form `N` or `N << M`.
func evalShift(t *testing.T, file, name, expr string) int64 {
	t.Helper()
	expr = strings.TrimSpace(expr)
	if !strings.Contains(expr, "<<") {
		v, err := strconv.ParseInt(expr, 10, 64)
		if err != nil {
			t.Fatalf("parse %s=%q in %s: %v", name, expr, file, err)
		}
		return v
	}
	parts := strings.SplitN(expr, "<<", 2)
	base, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		t.Fatalf("parse base of %s=%q in %s: %v", name, expr, file, err)
	}
	shift, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		t.Fatalf("parse shift of %s=%q in %s: %v", name, expr, file, err)
	}
	return base << uint(shift)
}
