// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type FilesystemEventType string

const (
	FilesystemEventCreate FilesystemEventType = "CREATE"
	FilesystemEventWrite  FilesystemEventType = "WRITE"
	FilesystemEventDelete FilesystemEventType = "DELETE"
	FilesystemEventRename FilesystemEventType = "RENAME"
	FilesystemEventChmod  FilesystemEventType = "CHMOD"
)

type FilesystemEvent struct {
	Type      FilesystemEventType `json:"type"`
	Name      string              `json:"name"`
	IsDir     bool                `json:"isDir"`
	Timestamp int64               `json:"timestamp"`
} // @name FilesystemEvent

type FileWatcher struct {
	path      string
	recursive bool
	events    chan FilesystemEvent
	errors    chan error
	watcher   *fsnotify.Watcher
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	running   bool
}

type WatchOptions struct {
	Recursive bool `json:"recursive"`
} // @name WatchOptions
