// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Constants for WebSocket and execution management.
const (
	writeWait         = 10 * time.Second
	gracePeriod       = 2 * time.Second
	workerScriptPerms = 0o700
	// createAckTimeout bounds how long TSFactory.Create waits for the host's
	// "created"/error acknowledgment so a stalled host can't hang Create forever.
	createAckTimeout = 15 * time.Second
	// maxWorkerLineBytes bounds a single newline-delimited worker output frame the
	// readers will buffer. Set generously above the interpreters' emit-side cap
	// (MAX_CHUNK_BYTES, 1 MiB) so legitimate frames never trip it, while still
	// preventing a malformed/runaway worker from growing reader memory without
	// bound ([B1]). An over-long line is drained and reported as an error chunk to
	// the affected context rather than crashing the reader ([C3]).
	maxWorkerLineBytes = 8 * 1024 * 1024
)

// WebSocket close codes (4000-4999 are for private/application use).
const (
	WebSocketCloseTimeout = 4008 // Execution timeout
)

// Chunk types in the daemon's WebSocket protocol.
const (
	ChunkTypeStdout  = "stdout"
	ChunkTypeStderr  = "stderr"
	ChunkTypeError   = "error"
	ChunkTypeDisplay = "display"
	ChunkTypeControl = "control"
)

// Control chunk subtypes (carried in OutputMessage.Text).
const (
	ControlChunkTypeCompleted   = "completed"
	ControlChunkTypeInterrupted = "interrupted"
)

// Command execution statuses (used internally to track an in-flight exec).
const (
	CommandStatusRunning = "running"
	CommandStatusOK      = "ok"
	CommandStatusError   = "error"
	CommandStatusTimeout = "timeout"
)

// Languages supported by the session daemon.
const (
	LanguagePython     = "python"
	LanguageTypeScript = "typescript"
	LanguageJavaScript = "javascript"
	LanguageBash       = "bash"
)

// CreateSessionRequest is the body of POST /sessions.
// The id is supplied by the API server (the same uuid stored in session.id);
// the daemon does NOT generate ids of its own. This collapses identity to one place.
type CreateSessionRequest struct {
	ID            string `json:"id" binding:"required"`
	Language      string `json:"language" binding:"required"`
	Cwd           string `json:"cwd"`
	MemoryLimitMB int    `json:"memoryLimitMb"`
}

// ExecuteRequest is the first frame on the /sessions/:id/execute WebSocket.
type ExecuteRequest struct {
	Code    string            `json:"code" binding:"required"`
	Envs    map[string]string `json:"envs"`
	Timeout *int64            `json:"timeout"` // seconds, 0 = no timeout
	// Reset asks the worker to drop any prior in-context state (Python: rebuild
	// globals; TypeScript: recreate the inner V8 context + per-session module
	// cache) BEFORE executing this frame's code. Used by the API's "transient"
	// context optimisation: keep one warm worker per (instance, language) and
	// recycle it across one-shot calls without leaking state. Persistent
	// contexts always omit this flag (so REPL semantics are preserved).
	Reset bool `json:"reset,omitempty"`
}

// SessionInfo is the public projection of a context (returned by GET /sessions).
type SessionInfo struct {
	ID         string    `json:"id"`
	Language   string    `json:"language"`
	Cwd        string    `json:"cwd"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	Active     bool      `json:"active"`
}

// PackageInfo is one entry in GET /packages.
type PackageInfo struct {
	Name              string `json:"name"`
	Version           string `json:"version,omitempty"`
	HasNativeBindings bool   `json:"hasNativeBindings"`
}

// OutputMessage is one frame in the daemon's WebSocket output protocol.
// Identical shape across Python and TypeScript engines so SDK code can
// be written against one frame format.
type OutputMessage struct {
	Type      string            `json:"type"`
	Text      string            `json:"text,omitempty"`
	Name      string            `json:"name,omitempty"`
	Value     string            `json:"value,omitempty"`
	Traceback string            `json:"traceback,omitempty"`
	Formats   []string          `json:"formats,omitempty"`
	Data      map[string]string `json:"data,omitempty"` // mime -> payload (base64 for binary)
}

// CommandExecution tracks a single in-flight code execution.
type CommandExecution struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	Status    string     `json:"status"`
	Error     *Error     `json:"error,omitempty"`
	StartedAt time.Time  `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
}

// Error is the structured error payload emitted on the WebSocket.
type Error struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Traceback string `json:"traceback,omitempty"`
}

// WorkerCommand is the JSON-line protocol message sent to a worker over stdin.
type WorkerCommand struct {
	Op            string            `json:"op"`
	SessionID     string            `json:"sessionId,omitempty"`
	ID            string            `json:"id,omitempty"`
	Code          string            `json:"code,omitempty"`
	Envs          map[string]string `json:"envs,omitempty"`
	MemoryLimitMB int               `json:"memoryLimitMb,omitempty"`
	ExecTimeoutMS int64             `json:"execTimeoutMs,omitempty"`
	Cwd           string            `json:"cwd,omitempty"`
	Language      string            `json:"language,omitempty"`
	// Reset, when true, asks the worker to wipe per-context state before
	// executing this command. See ExecuteRequest.Reset for the user-facing
	// rationale (transient-context recycling for one-shot codeRun).
	Reset bool `json:"reset,omitempty"`
	// Reply is an opaque correlation token echoed back by the host on reply chunks
	// (currently used for the list-packages round-trip). The host returns it on
	// the matching control chunk so the factory can route to the right pending
	// channel; without it the reply table never resolves and callers time out.
	Reply string `json:"reply,omitempty"`
}

// WorkerChunk is a JSON-line message read back from a worker over stdout.
// Same shape as OutputMessage but with an optional sessionId for multiplexed hosts (TS).
type WorkerChunk struct {
	SessionID string            `json:"sessionId,omitempty"`
	Type      string            `json:"type"`
	Text      string            `json:"text,omitempty"`
	Name      string            `json:"name,omitempty"`
	Value     string            `json:"value,omitempty"`
	Traceback string            `json:"traceback,omitempty"`
	Formats   []string          `json:"formats,omitempty"`
	Data      map[string]string `json:"data,omitempty"`
	Packages  []PackageInfo     `json:"packages,omitempty"`
	Reply     string            `json:"reply,omitempty"` // request-reply id (e.g., "list-packages-<n>")
	// Bash one-shot call result fields (carried on the "bash-call-result"
	// control chunk used by the TS/Python bash() bridges).
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exitCode,omitempty"`
}

// closeRequest is an internal frame asking the WS writer to send a close message.
type closeRequest struct {
	code    int
	message string
}

// wsFrame is the unit pushed to the WS writer goroutine.
type wsFrame struct {
	output *OutputMessage
	close  *closeRequest
}

// execJob is one queued execution waiting for its turn on the worker.
type execJob struct {
	code    string
	envs    map[string]string
	timeout time.Duration
	reset   bool   // ExecuteRequest.Reset — see types.go for rationale
	wsRefID string // identifies the attached WebSocket; unused beyond logging today
	doneCh  chan execResult
}

type execResult struct {
	cmd *CommandExecution
	Err error
}

// Session is one logical execution context. It is a thin façade around a worker
// (process or shared-host slot) plus a FIFO queue and an optional WebSocket sink.
type Session struct {
	info SessionInfo

	worker Worker
	mu     sync.Mutex

	queueOnce sync.Once
	queue     chan execJob
	queueCtx  context.Context
	queueStop context.CancelFunc
	// shuttingDown is set (under mu) by shutdown() before it cancels queueCtx and
	// launches drainAndClose. Enqueue checks it under the same lock so no job can
	// be reserved/sent once teardown has begun — see Enqueue / drainAndClose.
	shuttingDown bool

	activeCommand *CommandExecution
	commandMu     sync.Mutex

	// busy is the number of in-flight execs on this context (0 or 1 in practice — the
	// FIFO queue serializes execs per context). Tracked separately from activeCommand so
	// the /load endpoint can count "busy" contexts without taking commandMu under load.
	busy atomic.Int64

	// inflight counts execs Enqueued but not yet completed (queued OR running).
	// Incremented before the queue send so the idle sweeper cannot delete a context
	// out from under a just-accepted job — busy and LastUsedAt only update once the
	// queue goroutine starts runJob, which can lag the Enqueue.
	inflight atomic.Int64

	// Currently attached WebSocket client (only one at a time).
	client *wsClient

	// logger is used for best-effort teardown diagnostics (e.g. worker.Shutdown
	// errors). May be nil for sessions constructed without a manager logger.
	logger *slog.Logger
}
