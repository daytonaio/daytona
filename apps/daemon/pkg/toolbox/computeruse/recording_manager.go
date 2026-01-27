// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var (
	ErrRecordingNotFound    = errors.New("recording not found")
	ErrRecordingNotActive   = errors.New("recording is not active")
	ErrRecordingStillActive = errors.New("cannot delete an active recording")
	ErrFFmpegNotFound       = errors.New("ffmpeg not found in PATH")
	ErrNoDisplay            = errors.New("DISPLAY environment variable not set")
)

// activeRecording holds the state of a currently running recording
type activeRecording struct {
	recording  *Recording
	cmd        *exec.Cmd
	stdinPipe  io.WriteCloser
	cancelFunc func()
}

// RecordingManager manages screen recording sessions
type RecordingManager struct {
	mu               sync.RWMutex
	activeRecordings map[string]*activeRecording
	recordingsDir    string
}

// Global recording manager instance
var (
	globalRecordingManager *RecordingManager
	recordingManagerOnce   sync.Once
)

// GetRecordingManager returns the singleton RecordingManager instance
func GetRecordingManager() *RecordingManager {
	recordingManagerOnce.Do(func() {
		// Default recordings directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/tmp"
		}
		recordingsDir := filepath.Join(homeDir, ".daytona", "recordings")
		globalRecordingManager = &RecordingManager{
			activeRecordings: make(map[string]*activeRecording),
			recordingsDir:    recordingsDir,
		}
	})
	return globalRecordingManager
}

// GetRecordingsDir returns the recordings directory path
func (m *RecordingManager) GetRecordingsDir() string {
	return m.recordingsDir
}

// StartRecording starts a new screen recording session
func (m *RecordingManager) StartRecording(label string) (*Recording, error) {
	// Ensure recordings directory exists
	if err := os.MkdirAll(m.recordingsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create recordings directory: %w", err)
	}

	// Check if ffmpeg is available
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, ErrFFmpegNotFound
	}

	// Check for DISPLAY environment variable (required for X11)
	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0" // Default to :0 if not set
	}

	// Generate recording ID and filename
	// ID is included in filename so it can be recovered when scanning disk
	id := uuid.New().String()
	now := time.Now()
	timestamp := now.Format("20060102_150405")

	var fileName string
	if label != "" {
		fileName = fmt.Sprintf("%s_%s_%s.mp4", id, label, timestamp)
	} else {
		fileName = fmt.Sprintf("%s_session_%s.mp4", id, timestamp)
	}

	filePath := filepath.Join(m.recordingsDir, fileName)

	// Create recording entry
	recording := &Recording{
		ID:        id,
		FileName:  fileName,
		FilePath:  filePath,
		StartTime: now,
		Status:    "recording",
	}

	// Build ffmpeg command for Linux screen capture using x11grab
	// -f x11grab: X11 screen capture
	// -framerate 30: 30 FPS
	// -i :0.0: Capture from display :0, screen 0
	// -c:v libx264: H.264 codec
	// -preset ultrafast: Fast encoding for real-time capture
	// -pix_fmt yuv420p: Standard pixel format for compatibility
	cmd := exec.Command(ffmpegPath,
		"-f", "x11grab",
		"-framerate", "30",
		"-i", display,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-y", // Overwrite output file if exists
		filePath,
	)

	// Set environment to ensure DISPLAY is available
	cmd.Env = append(os.Environ(), fmt.Sprintf("DISPLAY=%s", display))

	// Get stdin pipe for graceful shutdown
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Start ffmpeg process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	log.Infof("Started recording %s to %s (DISPLAY=%s)", id, filePath, display)

	// Store active recording
	m.mu.Lock()
	m.activeRecordings[id] = &activeRecording{
		recording: recording,
		cmd:       cmd,
		stdinPipe: stdinPipe,
	}
	m.mu.Unlock()

	// Start a goroutine to wait for the process and handle unexpected exits
	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		defer m.mu.Unlock()

		if active, exists := m.activeRecordings[id]; exists {
			if err != nil {
				log.Warnf("Recording %s ffmpeg process exited with error: %v", id, err)
				active.recording.Status = "failed"
			}
			// Clean up from active recordings if still there
			delete(m.activeRecordings, id)
		}
	}()

	return recording, nil
}

// StopRecording stops an active recording session
func (m *RecordingManager) StopRecording(id string) (*Recording, error) {
	m.mu.Lock()
	active, exists := m.activeRecordings[id]
	if !exists {
		m.mu.Unlock()
		return nil, ErrRecordingNotFound
	}

	// Remove from active recordings
	delete(m.activeRecordings, id)
	m.mu.Unlock()

	// Send 'q' to ffmpeg stdin for graceful shutdown
	// This allows ffmpeg to properly finalize the video file
	if active.stdinPipe != nil {
		_, err := active.stdinPipe.Write([]byte("q"))
		if err != nil {
			log.Warnf("Failed to send quit signal to ffmpeg: %v", err)
		}
		active.stdinPipe.Close()
	}

	// Wait for ffmpeg to finish (with timeout)
	done := make(chan error, 1)
	go func() {
		done <- active.cmd.Wait()
	}()

	select {
	case <-done:
		// Process exited normally
	case <-time.After(10 * time.Second):
		// Force kill if it doesn't exit gracefully
		log.Warnf("Recording %s did not stop gracefully, force killing", id)
		if active.cmd.Process != nil {
			active.cmd.Process.Kill()
		}
	}

	// Update recording metadata
	now := time.Now()
	active.recording.EndTime = &now
	active.recording.Status = "completed"

	duration := now.Sub(active.recording.StartTime).Seconds()
	active.recording.DurationSeconds = &duration

	// Get file size
	if fileInfo, err := os.Stat(active.recording.FilePath); err == nil {
		size := fileInfo.Size()
		active.recording.SizeBytes = &size
	}

	log.Infof("Stopped recording %s, duration: %.2f seconds", id, duration)

	return active.recording, nil
}

// GetRecording returns a recording by ID (active or from filesystem)
func (m *RecordingManager) GetRecording(id string) (*Recording, error) {
	// First check active recordings
	m.mu.RLock()
	if active, exists := m.activeRecordings[id]; exists {
		recording := *active.recording
		m.mu.RUnlock()
		return &recording, nil
	}
	m.mu.RUnlock()

	// Search in completed recordings on disk
	recordings, err := m.ListRecordings()
	if err != nil {
		return nil, err
	}

	for _, rec := range recordings {
		if rec.ID == id {
			return &rec, nil
		}
	}

	return nil, ErrRecordingNotFound
}

// ListRecordings returns all recordings (active and completed)
func (m *RecordingManager) ListRecordings() ([]Recording, error) {
	recordings := []Recording{}

	// Add active recordings
	m.mu.RLock()
	for _, active := range m.activeRecordings {
		recordings = append(recordings, *active.recording)
	}
	m.mu.RUnlock()

	// Scan recordings directory for completed recordings
	if _, err := os.Stat(m.recordingsDir); os.IsNotExist(err) {
		return recordings, nil
	}

	entries, err := os.ReadDir(m.recordingsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read recordings directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only include MP4 files
		if filepath.Ext(entry.Name()) != ".mp4" {
			continue
		}

		// Skip files that are currently being recorded
		isActive := false
		m.mu.RLock()
		for _, active := range m.activeRecordings {
			if active.recording.FileName == entry.Name() {
				isActive = true
				break
			}
		}
		m.mu.RUnlock()

		if isActive {
			continue
		}

		filePath := filepath.Join(m.recordingsDir, entry.Name())
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		// Extract ID from filename (format: {id}_{label}_{timestamp}.mp4 or {id}_session_{timestamp}.mp4)
		// The ID is the first part before the underscore (UUID format)
		fileName := entry.Name()
		var recordingID string
		if idx := strings.Index(fileName, "_"); idx > 0 {
			potentialID := fileName[:idx]
			// Validate it's a UUID
			if _, err := uuid.Parse(potentialID); err == nil {
				recordingID = potentialID
			}
		}
		// Fallback to generating ID from file path for legacy recordings without ID in filename
		if recordingID == "" {
			recordingID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(filePath)).String()
		}

		// Create recording entry from file info
		// Use file modification time as a proxy for end time
		modTime := fileInfo.ModTime()
		size := fileInfo.Size()

		recording := Recording{
			ID:        recordingID,
			FileName:  fileName,
			FilePath:  filePath,
			StartTime: modTime, // Approximation - actual start time unknown for old recordings
			EndTime:   &modTime,
			Status:    "completed",
			SizeBytes: &size,
		}

		recordings = append(recordings, recording)
	}

	return recordings, nil
}

// DeleteRecording deletes a recording by ID
func (m *RecordingManager) DeleteRecording(id string) error {
	// Check if it's an active recording
	m.mu.RLock()
	if _, exists := m.activeRecordings[id]; exists {
		m.mu.RUnlock()
		return ErrRecordingStillActive
	}
	m.mu.RUnlock()

	// Find the recording
	recording, err := m.GetRecording(id)
	if err != nil {
		return err
	}

	// Delete the file
	if err := os.Remove(recording.FilePath); err != nil {
		if os.IsNotExist(err) {
			return ErrRecordingNotFound
		}
		return fmt.Errorf("failed to delete recording file: %w", err)
	}

	log.Infof("Deleted recording %s at %s", id, recording.FilePath)
	return nil
}
