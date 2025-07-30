// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// NewFileWatcher creates a new file watcher for the specified path
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

	// Check if path exists
	if _, err := os.Stat(w.path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", w.path)
	}

	// Build inotifywait command
	args := []string{
		"-m",                              // Monitor continuously
		"-e", "create,delete,modify,move", // Events to watch
		"--format", "%w%f|%e|%T", // Output format: path|event|timestamp
		"--timefmt", "%s", // Timestamp format (Unix timestamp)
	}

	if w.recursive {
		args = append(args, "-r") // Recursive
	}

	args = append(args, w.path) // Path to watch

	// Create command with context for cancellation
	w.cmd = exec.CommandContext(w.ctx, "inotifywait", args...)

	stdout, err := w.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := w.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := w.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start inotifywait: %w", err)
	}

	w.running = true

	// Handle stderr (errors and initial setup messages)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// inotifywait writes setup messages to stderr, only log actual errors
			if strings.Contains(line, "No space left on device") ||
				strings.Contains(line, "permission denied") ||
				strings.Contains(line, "failed") {
				log.Errorf("inotifywait error: %s", line)
				w.errors <- fmt.Errorf("inotifywait error: %s", line)
			}
		}
	}()

	// Process stdout (actual events)
	go func() {
		defer func() {
			w.mu.Lock()
			w.running = false
			w.mu.Unlock()
			close(w.events)
			close(w.errors)
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if event, err := w.parseInotifyOutput(line); err == nil {
				select {
				case w.events <- event:
				case <-w.ctx.Done():
					return
				}
			} else {
				log.Errorf("Failed to parse inotify output: %s - %v", line, err)
			}
		}

		if err := scanner.Err(); err != nil {
			w.errors <- fmt.Errorf("error reading inotifywait output: %w", err)
		}
	}()

	// Wait for command to finish in a separate goroutine
	go func() {
		err := w.cmd.Wait()
		if err != nil && w.ctx.Err() == nil {
			// Only report error if not cancelled
			w.errors <- fmt.Errorf("inotifywait process exited: %w", err)
		}
	}()

	return nil
}

// Stop stops the file watcher
func (w *FileWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	// Cancel context to stop goroutines
	w.cancel()

	// Kill the inotifywait process if it's still running
	if w.cmd != nil && w.cmd.Process != nil {
		w.cmd.Process.Kill()
	}
}

// Events returns the channel for receiving file system events
func (w *FileWatcher) Events() <-chan FilesystemEvent {
	return w.events
}

// Errors returns the channel for receiving errors
func (w *FileWatcher) Errors() <-chan error {
	return w.errors
}

// parseInotifyOutput parses a line of inotifywait output into a FilesystemEvent
// Format: "/path/to/file|CREATE,ISDIR|1640995200"
func (w *FileWatcher) parseInotifyOutput(line string) (FilesystemEvent, error) {
	parts := strings.Split(line, "|")
	if len(parts) != 3 {
		return FilesystemEvent{}, fmt.Errorf("invalid inotify output format: %s", line)
	}

	filePath := parts[0]
	eventTypes := parts[1]
	timestampStr := parts[2]

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		timestamp = time.Now().Unix()
	}

	// Determine if it's a directory
	isDir := strings.Contains(eventTypes, "ISDIR")

	// Convert inotify events to our event types
	var eventType FilesystemEventType
	switch {
	case strings.Contains(eventTypes, "CREATE"):
		eventType = FilesystemEventCreate
	case strings.Contains(eventTypes, "DELETE"):
		eventType = FilesystemEventDelete
	case strings.Contains(eventTypes, "MODIFY"):
		eventType = FilesystemEventWrite
	case strings.Contains(eventTypes, "MOVE"):
		eventType = FilesystemEventRename
	default:
		// Default to WRITE for any other modification
		eventType = FilesystemEventWrite
	}

	// Convert absolute path to relative path if possible
	relPath := filePath
	if absPath, err := filepath.Abs(w.path); err == nil {
		if rel, err := filepath.Rel(absPath, filePath); err == nil && !strings.HasPrefix(rel, "..") {
			relPath = filepath.Join(w.path, rel)
		}
	}

	return FilesystemEvent{
		Type:      eventType,
		Name:      relPath,
		IsDir:     isDir,
		Timestamp: timestamp,
	}, nil
}
