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

// Constants for WebSocket and execution management
const (
	writeWait         = 10 * time.Second
	gracePeriod       = 2 * time.Second
	workerScriptPerms = 0700
)

// WebSocket close codes (4000-4999 are for private/application use)
const (
	WebSocketCloseTimeout = 4008 // Execution timeout
)

// Chunk type constants for websocket streaming
const (
	ChunkTypeStdout  = "stdout"
	ChunkTypeStderr  = "stderr"
	ChunkTypeError   = "error"
	ChunkTypeControl = "control"
)

// Control chunk subtypes
const (
	ControlChunkTypeCompleted   = "completed"
	ControlChunkTypeInterrupted = "interrupted"
)

// Command execution statuses
const (
	CommandStatusRunning = "running"
	CommandStatusOK      = "ok"
	CommandStatusError   = "error"
	CommandStatusTimeout = "timeout"
)

// Supported languages
const (
	LanguagePython = "python"
)

// Controller handles interpreter-related HTTP endpoints
type Controller struct {
	workDir string
}

// API Request/Response types

// CreateContextRequest represents a request to create a new interpreter context
type CreateContextRequest struct {
	Cwd      *string `json:"cwd" validate:"optional"`
	Language *string `json:"language" validate:"optional"`
} // @name CreateContextRequest

// ExecuteRequest represents a request to execute code
type ExecuteRequest struct {
	Code      string             `json:"code" binding:"required"`
	ContextID *string            `json:"contextId" validate:"optional"`
	Timeout   *int64             `json:"timeout" validate:"optional"` // seconds, 0 disables timeout
	Envs      *map[string]string `json:"envs" validate:"optional"`
} // @name ExecuteRequest

// ListContextsResponse represents the response when listing contexts
type ListContextsResponse struct {
	Contexts []ContextInfo `json:"contexts"`
} // @name ListContextsResponse

// Context types

// ContextInfo contains metadata about an interpreter context
type ContextInfo struct {
	ID        string    `json:"id"`
	Cwd       string    `json:"cwd"`
	CreatedAt time.Time `json:"createdAt"`
	Active    bool      `json:"active"`
	Language  string    `json:"language"`
} // @name InterpreterContext

// Context represents an active interpreter context with operational methods
type Context struct {
	info ContextInfo

	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	ctx        context.Context
	cancel     context.CancelFunc
	workerPath string

	// Single websocket client (protected by mu)
	client *wsClient

	// Command tracking
	activeCommand *CommandExecution
	commandMu     sync.Mutex

	// Execution FIFO queue
	queue chan execJob

	// Process exit notification
	done chan struct{}

	// Guards session state and client
	mu sync.Mutex
}

// CommandExecution tracks a single code execution
type CommandExecution struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	Status    string     `json:"status"` // "running", "ok", "error", "interrupted", "exit", "timeout"
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

// Internal types

// wsClient represents a WebSocket client connection for output streaming
type wsClient struct {
	id        string
	conn      *websocket.Conn
	send      chan *OutputMessage
	done      chan struct{} // signals when clientWriter exits
	closeOnce sync.Once
}

// OutputMessage represents output sent to WebSocket clients
type OutputMessage struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Name      string `json:"name,omitempty"`
	Value     string `json:"value,omitempty"`
	Traceback string `json:"traceback,omitempty"`
}

// WorkerCommand represents a command sent to the language worker
type WorkerCommand struct {
	ID   string            `json:"id"`
	Code string            `json:"code,omitempty"`
	Envs map[string]string `json:"envs,omitempty"`
}

// execJob represents one queued execution
type execJob struct {
	code    string
	envs    map[string]string
	timeout time.Duration
	ws      *websocket.Conn
}
