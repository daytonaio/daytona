// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

func NewFileWatcher(path string, recursive bool) *FileWatcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &FileWatcher{
		path:      path,
		recursive: recursive,
		events:    make(chan FilesystemEvent, 100), // Buffer events
		errors:    make(chan error, 10),
		ctx:       ctx,
		cancel:    cancel,
		running:   false,
	}
}

// Start begins watching the specified path for file system changes
func (w *FileWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("file watcher is already running")
	}

	if _, err := os.Stat(w.path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", w.path)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	w.watcher = watcher
	w.running = true

	if err := w.addPath(w.path); err != nil {
		w.watcher.Close()
		w.running = false
		return fmt.Errorf("failed to add path to watcher: %w", err)
	}

	go w.processEvents()

	return nil
}

// addPath adds a path to the watcher, recursively if needed
func (w *FileWatcher) addPath(path string) error {
	err := w.watcher.Add(path)
	if err != nil {
		return err
	}

	if w.recursive {
		return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && walkPath != path {
				if err := w.watcher.Add(walkPath); err != nil {
					log.Warnf("Failed to watch directory %s: %v", walkPath, err)
				}
			}
			return nil
		})
	}

	return nil
}

// processEvents processes fsnotify events and converts them to our format
func (w *FileWatcher) processEvents() {
	defer func() {
		w.mu.Lock()
		w.running = false
		w.mu.Unlock()
		close(w.events)
		close(w.errors)
		if w.watcher != nil {
			w.watcher.Close()
		}
	}()

	for {
		select {
		case <-w.ctx.Done():
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleFsnotifyEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.errors <- err
		}
	}
}

// handleFsnotifyEvent converts fsnotify events to our event format
func (w *FileWatcher) handleFsnotifyEvent(event fsnotify.Event) {
	isDir := false
	if stat, err := os.Stat(event.Name); err == nil {
		isDir = stat.IsDir()
	}

	var eventType FilesystemEventType

	// Handle events in logical priority: Remove > Create > Rename > Write > Chmod
	// Rationale: Final state matters most (deleted > created > moved > modified > permissions)
	if event.Has(fsnotify.Remove) {
		eventType = FilesystemEventDelete
	} else if event.Has(fsnotify.Create) {
		eventType = FilesystemEventCreate
		if w.recursive && isDir {
			if err := w.watcher.Add(event.Name); err != nil {
				log.Warnf("Failed to watch new directory %s: %v", event.Name, err)
				// Don't send the event if we can't watch the directory
				return
			}
		}
	} else if event.Has(fsnotify.Rename) {
		eventType = FilesystemEventRename
	} else if event.Has(fsnotify.Write) {
		eventType = FilesystemEventWrite
	} else if event.Has(fsnotify.Chmod) {
		eventType = FilesystemEventChmod
	} else {
		return
	}

	filesystemEvent := FilesystemEvent{
		Type:      eventType,
		Name:      event.Name,
		IsDir:     isDir,
		Timestamp: time.Now().Unix(),
	}

	// Check context first to avoid race condition
	select {
	case <-w.ctx.Done():
		return
	default:
		// Try to send event, but don't block
		select {
		case w.events <- filesystemEvent:
		case <-w.ctx.Done():
			return
		default:
			// Channel is full, drop event to avoid blocking
			log.Warnf("Event channel full, dropping event for %s", event.Name)
		}
	}
}

// Stop stops the file watcher
func (w *FileWatcher) Stop() {
	w.cancel()
}

// Events returns the events channel
func (w *FileWatcher) Events() <-chan FilesystemEvent {
	return w.events
}

// Errors returns the errors channel
func (w *FileWatcher) Errors() <-chan error {
	return w.errors
}

// IsRunning returns whether the watcher is currently running
func (w *FileWatcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}
