// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"context"
	"io"
	"log/slog"
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
	logger  *slog.Logger
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
	Contexts []ContextInfo `json:"contexts" binding:"required"`
} // @name ListContextsResponse

// Context types

// ContextInfo contains metadata about an interpreter context
type ContextInfo struct {
	ID        string    `json:"id" binding:"required"`
	Cwd       string    `json:"cwd" binding:"required"`
	CreatedAt time.Time `json:"createdAt" binding:"required"`
	Active    bool      `json:"active" binding:"required"`
	Language  string    `json:"language" binding:"required"`
} // @name InterpreterContext

// Context represents an active interpreter context with operational methods
type Context struct {
	info ContextInfo

	logger *slog.Logger

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
	ID        string     `json:"id" binding:"required"`
	Code      string     `json:"code" binding:"required"`
	Status    string     `json:"status" binding:"required"` // "running", "ok", "error", "interrupted", "exit", "timeout"
	Error     *Error     `json:"error,omitempty"`
	StartedAt time.Time  `json:"startedAt" binding:"required"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
}

// Error represents a structured error from code execution
type Error struct {
	Name      string `json:"name" binding:"required"`
	Value     string `json:"value" binding:"required"`
	Traceback string `json:"traceback" binding:"required"`
}

// Internal types

// wsClient represents a WebSocket client connection for output streaming
type wsClient struct {
	id        string
	conn      *websocket.Conn
	send      chan wsFrame
	done      chan struct{} // signals when clientWriter exits
	closeOnce sync.Once
	logger    *slog.Logger
}

type wsFrame struct {
	output *OutputMessage
	close  *closeRequest
}

type closeRequest struct {
	code    int
	message string
}

// OutputMessage represents output sent to WebSocket clients
type OutputMessage struct {
	Type      string `json:"type" binding:"required"`
	Text      string `json:"text" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Value     string `json:"value" binding:"required"`
	Traceback string `json:"traceback" binding:"required"`
}

// WorkerCommand represents a command sent to the language worker
type WorkerCommand struct {
	ID   string            `json:"id" binding:"required"`
	Code string            `json:"code" binding:"required"`
	Envs map[string]string `json:"envs" binding:"required"`
}

// execJob represents one queued execution
type execJob struct {
	code    string
	envs    map[string]string
	timeout time.Duration
	ws      *websocket.Conn
}
