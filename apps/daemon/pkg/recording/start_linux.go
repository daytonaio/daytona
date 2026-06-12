//go:build linux

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/daytonaio/daemon/pkg/childreap"
	"github.com/google/uuid"
)

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

	s.logger.Debug("Started recording", "id", id, "path", filePath, "display", display)

	// Create a done channel to receive the Wait() result exactly once
	done := make(chan error, 1)

	// Store active recording
	active := &activeRecording{
		recording: recording,
		cmd:       cmd,
		stdinPipe: stdinPipe,
		done:      done,
	}
	s.activeRecordings.Set(id, active)

	// Start a goroutine to wait for the process and handle unexpected exits.
	// Reap (not Wait): we don't read stdout/stderr; ffmpeg output goes to a
	// file via -y target, so no Go I/O goroutines to drain.
	go func() {
		exitCode, err := childreap.Reap(cmd)

		// childreap.Reap returns (exitCode, nil) for *exec.ExitError cases —
		// unlike cmd.Wait, where non-zero exits surface as err != nil — so
		// check both to catch a crashed/killed ffmpeg (OOM, SIGKILL, corrupt
		// input, etc.).
		if err != nil || exitCode != 0 {
			// Unexpected exit (a graceful 'q' stop yields exit code 0 and the
			// entry is already popped by StopRecording). Keep the entry
			// visible as "failed" so list/stop/delete can report and clean it
			// up instead of having it silently vanish. markFailed must happen
			// before the done send so StopRecording observes the failure.
			if active, exists := s.activeRecordings.Get(id); exists {
				s.logger.Warn("Recording ffmpeg process exited unexpectedly",
					"id", id, "exitCode", exitCode, "error", err)
				active.markFailed(time.Now())
			}
		} else {
			// Clean exit outside StopRecording (e.g. external quit): drop the
			// active entry; the finalized file on disk represents it.
			s.activeRecordings.Pop(id)
		}

		done <- err // Signal the done channel with the result
	}()

	return recording, nil
}
