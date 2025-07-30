// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"os/exec"
	"sync"
)

// FilesystemEventType represents the type of file system event
type FilesystemEventType string

const (
	FilesystemEventCreate FilesystemEventType = "CREATE"
	FilesystemEventWrite  FilesystemEventType = "WRITE"
	FilesystemEventDelete FilesystemEventType = "DELETE"
	FilesystemEventRename FilesystemEventType = "RENAME"
)

// FilesystemEvent represents a file system change event
type FilesystemEvent struct {
	Type      FilesystemEventType `json:"type"`
	Name      string              `json:"name"`
	IsDir     bool                `json:"isDir"`
	Timestamp int64               `json:"timestamp"`
} // @name FilesystemEvent

// FileWatcher manages file watching using inotifywait
type FileWatcher struct {
	path      string
	recursive bool
	events    chan FilesystemEvent
	errors    chan error
	cmd       *exec.Cmd
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	running   bool
}

// WatchOptions represents options for file watching
type WatchOptions struct {
	Recursive bool `json:"recursive"`
} // @name WatchOptions
