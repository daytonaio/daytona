// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// Constants
const (
	writeWait = 10 * time.Second
	readLimit = 64 * 1024
)

// PTYController handles PTY-related HTTP endpoints
type PTYController struct {
	workDir string
}

// PTYManager manages multiple PTY sessions
type PTYManager struct {
	sessions cmap.ConcurrentMap[string, *PTYSession]
}

// wsClient represents a WebSocket client connection
type wsClient struct {
	id        string
	conn      *websocket.Conn
	send      chan []byte // outbound queue for this client (PTY -> WS)
	closeOnce sync.Once
}

// PTYSession represents a single PTY session with multi-client support
type PTYSession struct {
	info PTYSessionInfo

	cmd    *exec.Cmd
	ptmx   *os.File
	ctx    context.Context
	cancel context.CancelFunc

	// multi-attach
	clients   cmap.ConcurrentMap[string, *wsClient]
	clientsMu sync.RWMutex

	// funnel of all client inputs -> single PTY writer (preserves ordering)
	inCh chan []byte

	// guards general session fields (info/cmd/ptmx)
	mu sync.Mutex
}

// PTYSessionInfo contains metadata about a PTY session
type PTYSessionInfo struct {
	ID        string            `json:"id" validate:"required"`
	Cwd       string            `json:"cwd" validate:"required"`
	Envs      map[string]string `json:"envs" validate:"required"`
	Cols      uint16            `json:"cols" validate:"required"`
	Rows      uint16            `json:"rows" validate:"required"`
	CreatedAt time.Time         `json:"createdAt" validate:"required"`
	Active    bool              `json:"active" validate:"required"`
	LazyStart bool              `json:"lazyStart" validate:"required"` // Whether this session uses lazy start
} // @name PtySessionInfo

// API Request/Response types

// PTYCreateRequest represents a request to create a new PTY session
type PTYCreateRequest struct {
	ID        string            `json:"id"`
	Cwd       string            `json:"cwd,omitempty"`
	Envs      map[string]string `json:"envs,omitempty"`
	Cols      *uint16           `json:"cols" validate:"optional"`
	Rows      *uint16           `json:"rows" validate:"optional"`
	LazyStart bool              `json:"lazyStart,omitempty"` // Don't start PTY until first client connects
} // @name PtyCreateRequest

// PTYCreateResponse represents the response when creating a PTY session
type PTYCreateResponse struct {
	SessionID string `json:"sessionId" validate:"required"`
} // @name PtyCreateResponse

// PTYListResponse represents the response when listing PTY sessions
type PTYListResponse struct {
	Sessions []PTYSessionInfo `json:"sessions" validate:"required"`
} // @name PtyListResponse

// PTYResizeRequest represents a request to resize a PTY session
type PTYResizeRequest struct {
	Cols uint16 `json:"cols" binding:"required,min=1,max=1000"`
	Rows uint16 `json:"rows" binding:"required,min=1,max=1000"`
} // @name PtyResizeRequest
