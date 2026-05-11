// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeHostScript materializes an embedded interpreter script to disk so the Go
// process can exec it. It writes into a freshly-created PRIVATE subdirectory
// (os.MkdirTemp, mode 0700) of baseDir rather than directly into baseDir. The
// default baseDir is os.TempDir(), a world-writable directory; writing a
// deterministic filename there is a symlink/clobber TOCTOU — a local attacker can
// pre-create that path as a symlink, and os.WriteFile (O_CREATE|O_TRUNC) follows
// it, truncating the target. A per-process private subdir removes the predictable
// path and the shared-directory exposure. baseDir defaults to os.TempDir() when
// empty.
func writeHostScript(baseDir, filename, content string) (string, error) {
	if baseDir == "" {
		baseDir = os.TempDir()
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", fmt.Errorf("create host script base dir: %w", err)
	}
	dir, err := os.MkdirTemp(baseDir, "daytona-session-host-")
	if err != nil {
		return "", fmt.Errorf("create private host script dir: %w", err)
	}
	scriptPath := filepath.Join(dir, filename)
	if err := os.WriteFile(scriptPath, []byte(content), workerScriptPerms); err != nil {
		_ = os.RemoveAll(dir) // don't leak the just-created temp dir on the error path
		return "", fmt.Errorf("write host script: %w", err)
	}
	return scriptPath, nil
}

// removeHostScriptDir removes the private temp directory writeHostScript created
// for scriptPath. Factories call it on Shutdown so the per-process
// daytona-session-host-* directories don't accumulate across daemon restarts.
// Best-effort; a no-op for an empty path (factory built without a script).
func removeHostScriptDir(scriptPath string) {
	if scriptPath == "" {
		return
	}
	_ = os.RemoveAll(filepath.Dir(scriptPath))
}

// Worker is the contract every execution backend implements. A Worker owns
// the lifecycle of one logical context's execution slot — a subprocess for
// Python, a V8 session slot inside a shared host process for TypeScript.
//
// The Worker emits chunks via the supplied chunk-handler closure. Implementations
// may use one goroutine per worker (Python) or share one goroutine across many
// contexts (TS host); either way they MUST tag chunks with the right context id
// before invoking the handler.
type Worker interface {
	// Send queues a single command and returns immediately. The chunk handler
	// passed at construction time is invoked from a worker-owned goroutine for
	// each chunk. When the worker emits a {type:"control", text:"completed"|
	// "interrupted"} chunk, the caller knows the command is done.
	Send(cmd WorkerCommand) error

	// Interrupt asks the worker to abort the currently running command.
	// For subprocess workers this typically sends SIGINT then SIGKILL after
	// gracePeriod. For V8 session workers it disposes and recreates the session.
	Interrupt() error

	// Shutdown tears the worker down and returns any teardown error so the
	// caller can log it. After Shutdown returns, Send and Interrupt are no-ops
	// and the worker may be discarded.
	Shutdown() error

	// Active reports whether the worker is currently usable.
	Active() bool
}

// BashInvoker runs a single bash command on the shared just-bash host and
// returns the aggregated result. It backs the Python bash() bridge: the Python
// worker emits a hostcall over stdio, the daemon routes it here, and writes the
// result back. The TS bridge does not use this — it runs just-bash in its own
// Node host, in-process. Per-session state (overlay) is keyed by sessionID;
// Release drops it when the session ends.
type BashInvoker interface {
	Call(sessionID, code string, env map[string]string) (stdout, stderr string, exitCode int, err error)
	Release(sessionID string)
}
