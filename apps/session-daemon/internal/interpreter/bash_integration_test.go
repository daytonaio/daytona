// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build integration

// Integration tests that exercise the REAL host scripts (repl_bash_host.js,
// repl_host.js, repl_worker.py) against a real just-bash / node / python3.
//
// They are gated behind the `integration` build tag (not run in normal CI)
// because they require a node bundle with just-bash (and, for the TS bridge,
// isolated-vm + esbuild-wasm) installed. Point the bundle dir via
// DAYTONA_ITEST_BUNDLE (default: /tmp/daytona_itest_bundle).
//
// Set up the bundle once:
//
//	mkdir -p /tmp/daytona_itest_bundle && cd /tmp/daytona_itest_bundle && npm init -y
//	npm install just-bash@3.0.2 isolated-vm@5.0.3 esbuild-wasm@0.24.2
//
// Run:
//
//	go test -tags integration -run Integration -v ./internal/interpreter/...
package interpreter

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/daytonaio/session-daemon/internal/config"
)

func itestBundleRoot() string {
	if v := os.Getenv("DAYTONA_ITEST_BUNDLE"); v != "" {
		return v
	}
	return "/tmp/daytona_itest_bundle"
}

func itestConfig(t *testing.T) *config.Config {
	t.Helper()
	bundle := itestBundleRoot()
	if _, err := os.Stat(filepath.Join(bundle, "node_modules", "just-bash")); err != nil {
		t.Skipf("just-bash not found under %s/node_modules (set DAYTONA_ITEST_BUNDLE): %v", bundle, err)
	}
	tmp := t.TempDir()
	return &config.Config{
		WorkspaceRoot:      tmp,
		NodeBundleRoot:     bundle,
		PythonInterpreter:  "python3",
		NodeInterpreter:    "node",
		WorkerScriptDir:    tmp,
		HostScriptCacheDir: tmp,
		TSMaxContexts:      8,
		PyMaxContexts:      8,
		BashMaxContexts:    8,
		TSDefaultMemoryMB:  128,
		TSMaxMemoryMB:      512,
		PyWarmPoolSize:     0,
	}
}

func itestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// chunkCollector aggregates worker chunks and signals when a terminal
// (completed/interrupted) control chunk arrives. onChunk is invoked from a
// single worker-owned goroutine per worker, so a mutex is sufficient.
type chunkCollector struct {
	mu       sync.Mutex
	stdout   strings.Builder
	stderr   strings.Builder
	errName  string
	errValue string
	done     chan struct{}
	once     sync.Once
}

func newCollector() *chunkCollector { return &chunkCollector{done: make(chan struct{})} }

func (c *chunkCollector) onChunk(ch *WorkerChunk) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch ch.Type {
	case ChunkTypeStdout:
		c.stdout.WriteString(ch.Text)
	case ChunkTypeStderr:
		c.stderr.WriteString(ch.Text)
	case ChunkTypeError:
		c.errName, c.errValue = ch.Name, ch.Value
	case ChunkTypeControl:
		if ch.Text == ControlChunkTypeCompleted || ch.Text == ControlChunkTypeInterrupted {
			c.once.Do(func() { close(c.done) })
		}
	}
}

func (c *chunkCollector) wait(t *testing.T, d time.Duration) {
	t.Helper()
	select {
	case <-c.done:
	case <-time.After(d):
		t.Fatalf("timed out waiting for terminal chunk after %s; stdout=%q stderr=%q", d, c.out(), c.err())
	}
}

func (c *chunkCollector) out() string { c.mu.Lock(); defer c.mu.Unlock(); return c.stdout.String() }
func (c *chunkCollector) err() string { c.mu.Lock(); defer c.mu.Unlock(); return c.stderr.String() }

// TestIntegrationBashDirect runs a virtual-bash pipeline in a standalone bash
// isolate and asserts the streamed stdout.
func TestIntegrationBashDirect(t *testing.T) {
	cfg := itestConfig(t)
	bf, err := NewBashFactory(cfg, itestLogger())
	if err != nil {
		t.Fatalf("NewBashFactory: %v", err)
	}
	defer bf.Shutdown()

	col := newCollector()
	w, err := bf.Create("bash-1", CreateSessionRequest{ID: "bash-1", Language: LanguageBash}, col.onChunk)
	if err != nil {
		t.Fatalf("Create bash worker: %v", err)
	}
	defer func() { _ = w.Shutdown() }()

	if err := w.Send(WorkerCommand{Op: "exec", ID: "c1", Code: `echo "hello world" | grep -o hello`}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	col.wait(t, 20*time.Second)

	if got := col.out(); !strings.Contains(got, "hello") {
		t.Fatalf("bash stdout = %q, want it to contain %q", got, "hello")
	}
}

// TestIntegrationBashOverlayIsolation proves a write in one bash session's
// overlay is not visible to a second session (private, in-memory writes).
func TestIntegrationBashOverlayIsolation(t *testing.T) {
	cfg := itestConfig(t)
	bf, err := NewBashFactory(cfg, itestLogger())
	if err != nil {
		t.Fatalf("NewBashFactory: %v", err)
	}
	defer bf.Shutdown()

	// Session A writes into its overlay.
	colA := newCollector()
	wA, err := bf.Create("ov-a", CreateSessionRequest{ID: "ov-a", Language: LanguageBash}, colA.onChunk)
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	defer func() { _ = wA.Shutdown() }()
	probe := filepath.Join(cfg.WorkspaceRoot, "overlay_probe.txt")
	if err := wA.Send(WorkerCommand{Op: "exec", ID: "a1", Code: "echo secret > " + probe + " && cat " + probe}); err != nil {
		t.Fatalf("Send A: %v", err)
	}
	colA.wait(t, 20*time.Second)
	if got := colA.out(); !strings.Contains(got, "secret") {
		t.Fatalf("session A stdout = %q, want %q", got, "secret")
	}

	// Session B must not see A's overlay write.
	colB := newCollector()
	wB, err := bf.Create("ov-b", CreateSessionRequest{ID: "ov-b", Language: LanguageBash}, colB.onChunk)
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	defer func() { _ = wB.Shutdown() }()
	if err := wB.Send(WorkerCommand{Op: "exec", ID: "b1", Code: "cat " + probe + " 2>/dev/null || echo MISSING"}); err != nil {
		t.Fatalf("Send B: %v", err)
	}
	colB.wait(t, 20*time.Second)
	if got := colB.out(); !strings.Contains(got, "MISSING") {
		t.Fatalf("session B saw A's overlay write: stdout = %q, want %q", got, "MISSING")
	}

	// And it must not have leaked to the real filesystem either.
	if _, err := os.Stat(probe); err == nil {
		t.Fatalf("overlay write leaked to the real FS at %s", probe)
	}
}

// TestIntegrationPythonBashBridge proves Python user code can call bash() via
// the stdio hostcall RPC and read the result (stdout + exit code).
func TestIntegrationPythonBashBridge(t *testing.T) {
	cfg := itestConfig(t)
	bf, err := NewBashFactory(cfg, itestLogger())
	if err != nil {
		t.Fatalf("NewBashFactory: %v", err)
	}
	defer bf.Shutdown()

	pf, err := NewPythonFactory(cfg, itestLogger())
	if err != nil {
		t.Fatalf("NewPythonFactory: %v", err)
	}
	defer pf.Shutdown()
	pf.SetBashInvoker(bf)

	col := newCollector()
	w, err := pf.Create("py-1", CreateSessionRequest{ID: "py-1", Language: LanguagePython}, col.onChunk)
	if err != nil {
		t.Fatalf("Create python worker: %v", err)
	}
	defer func() { _ = w.Shutdown() }()

	code := strings.Join([]string{
		`r = bash("echo from-python | tr a-z A-Z")`,
		`print(r.stdout.strip())`,
		`print("exit", r.exit_code)`,
		`print("false-exit", bash("false").exit_code)`,
	}, "\n")
	if err := w.Send(WorkerCommand{Op: "exec", ID: "c1", Code: code}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	col.wait(t, 30*time.Second)

	out := col.out()
	for _, want := range []string{"FROM-PYTHON", "exit 0", "false-exit 1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("python bash() stdout = %q, want it to contain %q (stderr=%q)", out, want, col.err())
		}
	}
}

// TestIntegrationTypescriptBashBridge proves TS isolate code can call the
// host-bridged bash(). Skipped unless isolated-vm + esbuild-wasm are present in
// the bundle (they require a native build).
func TestIntegrationTypescriptBashBridge(t *testing.T) {
	cfg := itestConfig(t)
	for _, mod := range []string{"isolated-vm", "esbuild-wasm"} {
		if _, err := os.Stat(filepath.Join(cfg.NodeBundleRoot, "node_modules", mod)); err != nil {
			t.Skipf("%s not found in bundle; skipping TS bridge integration test", mod)
		}
	}

	tf, err := NewTSFactory(cfg, itestLogger())
	if err != nil {
		t.Fatalf("NewTSFactory: %v", err)
	}
	defer tf.Shutdown()

	col := newCollector()
	w, err := tf.Create("ts-1", CreateSessionRequest{ID: "ts-1", Language: LanguageTypeScript}, col.onChunk)
	if err != nil {
		t.Fatalf("Create ts worker: %v", err)
	}
	defer func() { _ = w.Shutdown() }()

	if err := w.Send(WorkerCommand{Op: "exec", ID: "c1", Code: `const r = await bash('echo from-ts'); console.log(r.stdout.trim(), 'exit', r.exitCode);`}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	col.wait(t, 60*time.Second)

	out := col.out()
	if !strings.Contains(out, "from-ts") || !strings.Contains(out, "exit 0") {
		t.Fatalf("ts bash() stdout = %q, want it to contain %q and %q (stderr=%q)", out, "from-ts", "exit 0", col.err())
	}
}
