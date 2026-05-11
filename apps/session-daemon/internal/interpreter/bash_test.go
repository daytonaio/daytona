// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/daytonaio/session-daemon/internal/config"
)

// stubFactory is a no-op WorkerFactory for wiring tests that never actually
// spawn a host (the real bash host needs Node + just-bash, which isn't present
// in unit-test CI). These tests exercise the routing/capacity logic only.
type stubFactory struct{}

func (stubFactory) Create(string, CreateSessionRequest, func(*WorkerChunk)) (Worker, error) {
	return nil, nil
}
func (stubFactory) ListPackages() ([]PackageInfo, error) { return nil, nil }
func (stubFactory) Shutdown()                            {}

func TestNormalizeLanguageBash(t *testing.T) {
	cases := map[string]string{
		"bash":       LanguageBash,
		"sh":         LanguageBash,
		"python":     LanguagePython,
		"":           LanguagePython,
		"javascript": LanguageTypeScript,
		"typescript": LanguageTypeScript,
	}
	for in, want := range cases {
		if got := normalizeLanguage(in); got != want {
			t.Errorf("normalizeLanguage(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFactoryForBash(t *testing.T) {
	m := &Manager{
		cfg:         &config.Config{},
		logger:      slog.Default(),
		bashFactory: stubFactory{},
		contexts:    map[string]*Session{},
	}
	for _, lang := range []string{"bash", "sh"} {
		f, err := m.factoryFor(lang)
		if err != nil {
			t.Fatalf("factoryFor(%q) error: %v", lang, err)
		}
		if f == nil {
			t.Fatalf("factoryFor(%q) returned nil factory", lang)
		}
	}

	// With no bash factory registered, bash is unsupported (the daemon booted
	// without just-bash) rather than silently routing elsewhere.
	mNoBash := &Manager{cfg: &config.Config{}, logger: slog.Default(), contexts: map[string]*Session{}}
	if _, err := mNoBash.factoryFor("bash"); !errors.Is(err, ErrUnsupportedLang) {
		t.Fatalf("factoryFor(bash) with nil factory = %v, want ErrUnsupportedLang", err)
	}
}

func TestCheckCapacityBash(t *testing.T) {
	m := &Manager{
		cfg:      &config.Config{PyMaxContexts: 10, TSMaxContexts: 10, BashMaxContexts: 2},
		logger:   slog.Default(),
		contexts: map[string]*Session{},
	}

	// Under cap: allowed.
	if err := m.checkCapacityLocked("bash"); err != nil {
		t.Fatalf("checkCapacityLocked(bash) under cap: %v", err)
	}

	// Fill bash to its cap; a third bash context must be refused while other
	// languages stay unaffected (per-language caps are independent).
	m.contexts["a"] = &Session{info: SessionInfo{Language: LanguageBash}}
	m.contexts["b"] = &Session{info: SessionInfo{Language: LanguageBash}}
	if err := m.checkCapacityLocked("bash"); !errors.Is(err, ErrCapacity) {
		t.Fatalf("checkCapacityLocked(bash) at cap = %v, want ErrCapacity", err)
	}
	if err := m.checkCapacityLocked("python"); err != nil {
		t.Fatalf("python capacity must be independent of bash: %v", err)
	}
	// "sh" normalizes to bash, so it shares the bash cap.
	if err := m.checkCapacityLocked("sh"); !errors.Is(err, ErrCapacity) {
		t.Fatalf("checkCapacityLocked(sh) at bash cap = %v, want ErrCapacity", err)
	}
}

func TestLoadCountsReportsBashMax(t *testing.T) {
	m := &Manager{
		cfg:      &config.Config{PyMaxContexts: 32, TSMaxContexts: 16, BashMaxContexts: 128},
		logger:   slog.Default(),
		contexts: map[string]*Session{},
	}
	_, _, _, _, bashMax := m.LoadCounts()
	if bashMax != 128 {
		t.Fatalf("LoadCounts bashMax = %d, want 128", bashMax)
	}
}

// TestHostCallRoundTrip locks the Python bash() bridge wire shape: the worker's
// hostcall request and the daemon's reply must (de)serialize with the fields
// repl_worker.py reads/writes.
func TestHostCallRoundTrip(t *testing.T) {
	req := hostCall{ID: "abc", Method: "bash", Cmd: "echo hi", Env: map[string]string{"K": "V"}}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal hostCall: %v", err)
	}
	var got hostCall
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal hostCall: %v", err)
	}
	if got.ID != "abc" || got.Method != "bash" || got.Cmd != "echo hi" || got.Env["K"] != "V" {
		t.Fatalf("hostCall round-trip mismatch: %+v", got)
	}

	res := hostCallResult{Type: "hostcall_result", ID: "abc", Stdout: "hi\n", Stderr: "", ExitCode: 0}
	data, err = json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal hostCallResult: %v", err)
	}
	// The Python side keys on type/id and reads stdout/stderr/exitCode.
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal hostCallResult: %v", err)
	}
	for _, k := range []string{"type", "id", "exitCode"} {
		if _, ok := m[k]; !ok {
			t.Errorf("hostCallResult JSON missing %q: %s", k, data)
		}
	}
}

// TestWorkerChunkBashCallFields ensures the bash-call result chunk emitted by
// repl_bash_host.js decodes into the fields BashFactory.Call reads.
func TestWorkerChunkBashCallFields(t *testing.T) {
	const line = `{"type":"control","text":"bash-call-result","reply":"bash-call-7","stdout":"out","stderr":"err","exitCode":2}`
	var c WorkerChunk
	if err := json.Unmarshal([]byte(line), &c); err != nil {
		t.Fatalf("unmarshal bash-call chunk: %v", err)
	}
	if c.Type != ChunkTypeControl || c.Text != "bash-call-result" || c.Reply != "bash-call-7" {
		t.Fatalf("routing fields mismatch: %+v", c)
	}
	if c.Stdout != "out" || c.Stderr != "err" || c.ExitCode != 2 {
		t.Fatalf("result fields mismatch: stdout=%q stderr=%q exit=%d", c.Stdout, c.Stderr, c.ExitCode)
	}
}
