// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"context"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Constants
const (
    writeWait         = 10 * time.Second
    readLimit         = 64 * 1024
    gracePeriod       = 2 * time.Second
    workerScriptPerms = 0700
)

// Chunk type constants for websocket streaming
const (
    ChunkTypeStdout   = "stdout"
    ChunkTypeStderr   = "stderr"
    ChunkTypeError    = "error"
    ChunkTypeArtifact = "artifact"
    ChunkTypeControl  = "control"
)

// InterpreterController handles interpreter-related HTTP endpoints
type InterpreterController struct {
	workDir string
}

// wsClient represents a WebSocket client connection for output streaming
type wsClient struct {
	id        string
	conn      *websocket.Conn
	send      chan *OutputMessage // outbound queue for this client
	closeOnce sync.Once
}

// InterpreterSession represents a single Python interpreter session
type InterpreterSession struct {
	info InterpreterSessionInfo

	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	ctx        context.Context
	cancel     context.CancelFunc
	workerPath string // path to the worker script

	// single optional websocket client
	client   *wsClient
	clientMu sync.RWMutex

	// command tracking
	activeCommand *CommandExecution
	commandMu     sync.Mutex

	// execution FIFO queue
	queue chan execJob

	// guards session state
	mu sync.Mutex
}

// InterpreterSessionInfo contains metadata about an interpreter session
type InterpreterSessionInfo struct {
	ID        string            `json:"id"`
	Cwd       string            `json:"cwd"`
	CreatedAt time.Time         `json:"createdAt"`
	Active    bool              `json:"active"`
	Language  string            `json:"language"` // "python" initially, can expand later
}

// CommandExecution tracks a single code execution (for internal state only - output is streamed)
type CommandExecution struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	Status    string     `json:"status"` // "running", "ok", "error", "interrupted", "exit"
	Error     *Error     `json:"error,omitempty"`
	StartedAt time.Time  `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
}

// Error represents a structured error from code execution
type Error struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Traceback string `json:"traceback,omitempty"`
}

// InterpreterExecuteRequest represents a request to execute code
type InterpreterExecuteRequest struct {
	Code    string            `json:"code" binding:"required"`
	Timeout *uint32           `json:"timeout,omitempty"` // timeout in seconds (0 means no timeout)
	Envs    map[string]string `json:"envs,omitempty"`
} // @name InterpreterExecuteRequest

// Internal message types for JSON protocol with Python worker

// WorkerCommand represents a command sent to the Python worker
type WorkerCommand struct {
	ID   string `json:"id"`
	Cmd  string `json:"cmd"`  // "exec" or "shutdown"
	Code string `json:"code,omitempty"`
    Envs map[string]string `json:"envs,omitempty"`
}

// Chunk represents a streaming chunk from the Python worker (new protocol)
type Chunk struct {
	Type      string         `json:"type"` // "stdout", "stderr", "error", "artifact", "control"
	Text      string         `json:"text,omitempty"`
	Name      string         `json:"name,omitempty"`      // For error chunks
	Value     string         `json:"value,omitempty"`     // For error chunks
	Traceback string         `json:"traceback,omitempty"` // For error chunks
	Artifact  map[string]any `json:"artifact,omitempty"`  // For artifact chunks
}

// OutputMessage represents output sent to WebSocket clients (same format as Chunk)
type OutputMessage struct {
	Type      string         `json:"type"` // "stdout", "stderr", "error", "artifact", "control"
	Text      string         `json:"text,omitempty"`
	Name      string         `json:"name,omitempty"`      // For error type
	Value     string         `json:"value,omitempty"`     // For error type
	Traceback string         `json:"traceback,omitempty"` // For error type
	Artifact  map[string]any `json:"artifact,omitempty"`  // For artifact type
}

// execJob represents one queued execution
type execJob struct {
    code    string
    envs    map[string]string
    timeout time.Duration
    ws      *websocket.Conn
}

