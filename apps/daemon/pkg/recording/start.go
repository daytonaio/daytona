// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// validateLabel validates a user-provided label to prevent path injection
// and ensure it's safe for use in a filename. Returns error if invalid.
func validateLabel(label string) error {
	const maxLabelLength = 100

	// Trim whitespace for validation
	trimmed := strings.TrimSpace(label)

	// Check if label is empty after trimming
	if trimmed == "" {
		return ErrInvalidLabel
	}

	// Check length
	if len(label) > maxLabelLength {
		return ErrInvalidLabel
	}

	// Check for path separators (directory traversal)
	if strings.Contains(label, "/") || strings.Contains(label, "\\") {
		return ErrInvalidLabel
	}

	// Check for leading dots (hidden files)
	if strings.HasPrefix(trimmed, ".") {
		return ErrInvalidLabel
	}

	// Only allow safe characters: alphanumeric, spaces, dots, underscores, and hyphens
	safePattern := regexp.MustCompile(`^[A-Za-z0-9.\s_-]+$`)
	if !safePattern.MatchString(label) {
		return ErrInvalidLabel
	}

	return nil
}

// StartRecording starts a new screen recording session
func (s *RecordingService) StartRecording(label *string) (*Recording, error) {
	// Ensure recordings directory exists
	if err := os.MkdirAll(s.recordingsDir, 0755); err != nil {
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

	// Validate label if provided (reject invalid labels without modification)
	if label != nil && *label != "" {
		if err := validateLabel(*label); err != nil {
			return nil, err
		}
	}

	var fileName string
	if label != nil && *label != "" {
		fileName = fmt.Sprintf("%s_%s_%s.mp4", id, *label, timestamp)
	} else {
		fileName = fmt.Sprintf("%s_session_%s.mp4", id, timestamp)
	}

	filePath := filepath.Join(s.recordingsDir, fileName)

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

	log.Debugf("Started recording %s to %s (DISPLAY=%s)", id, filePath, display)

	// Create a done channel to receive the Wait() result exactly once
	done := make(chan error, 1)

	// Store active recording
	s.activeRecordings.Set(id, &activeRecording{
		recording: recording,
		cmd:       cmd,
		stdinPipe: stdinPipe,
		done:      done,
	})

	// Start a goroutine to wait for the process and handle unexpected exits
	go func() {
		err := cmd.Wait()
		done <- err // Signal the done channel with the result

		// Atomically remove from active recordings if still there
		if active, exists := s.activeRecordings.Pop(id); exists {
			if err != nil {
				log.Warnf("Recording %s ffmpeg process exited unexpectedly with error: %v", id, err)
				active.recording.Status = "failed"
			}
		}
	}()

	return recording, nil
}
