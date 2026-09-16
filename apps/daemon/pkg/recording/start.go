// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/daytonaio/daemon/pkg/childreap"
	"github.com/google/uuid"
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
	// Validate label if provided (reject invalid labels without modification)
	if label != nil && *label != "" {
		if err := validateLabel(*label); err != nil {
			return nil, err
		}
	}

	// Ensure recordings directory exists
	if err := os.MkdirAll(s.recordingsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create recordings directory: %w", err)
	}

	// Check if ffmpeg is available
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, ErrFFmpegNotFound
	}

	// Generate recording ID and filename
	// ID is included in filename so it can be recovered when scanning disk
	id := uuid.New().String()
	now := time.Now()
	timestamp := now.Format("20060102_150405")

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

	// Build platform-specific ffmpeg capture arguments
	var captureArgs []string
	switch runtime.GOOS {
	case "windows":
		// Windows GDI screen capture
		// -f gdigrab: Windows GDI screen capture device
		// -framerate 30: 30 FPS
		// -i desktop: Capture the entire desktop
		captureArgs = []string{
			"-f", "gdigrab",
			"-framerate", "30",
			"-i", "desktop",
		}
	case "linux":
		// X11 screen capture
		// -f x11grab: X11 screen capture
		// -framerate 30: 30 FPS
		// -i :0: Capture from display :0, screen 0
		display := os.Getenv("DISPLAY")
		if display == "" {
			display = ":0" // Default to :0 if not set
		}
		captureArgs = []string{
			"-f", "x11grab",
			"-framerate", "30",
			"-i", display,
		}
	default:
		return nil, ErrUnsupportedPlatform
	}

	// Common encoding arguments
	// -c:v libx264: H.264 codec
	// -preset ultrafast: Fast encoding for real-time capture
	// -pix_fmt yuv420p: Standard pixel format for compatibility
	// -y: Overwrite output file if exists
	args := append(captureArgs,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-y",
		filePath,
	)

	cmd := exec.Command(ffmpegPath, args...)

	// On Linux, ensure DISPLAY environment variable is set for X11
	if runtime.GOOS == "linux" {
		display := os.Getenv("DISPLAY")
		if display == "" {
			display = ":0"
		}
		cmd.Env = append(os.Environ(), fmt.Sprintf("DISPLAY=%s", display))
	}

	// Get stdin pipe for graceful shutdown
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Start ffmpeg process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	s.logger.Debug("Started recording", "id", id, "path", filePath, "platform", runtime.GOOS)

	// Create a done channel to receive the Wait() result exactly once
	done := make(chan error, 1)

	// Start a goroutine to wait for the process and handle unexpected exits.
	// Reap (not Wait): we don't read stdout/stderr; ffmpeg output goes to a
	// file via -y target, so no Go I/O goroutines to drain.
	// NOTE: This goroutine must be started BEFORE the startup health check
	// so the done channel receives the exit status if ffmpeg crashes immediately.
	go func() {
		exitCode, err := childreap.Reap(cmd)
		done <- err // Signal the done channel with the result

		// Atomically remove from active recordings if still there
		if active, exists := s.activeRecordings.Pop(id); exists {
			// childreap.Reap returns (exitCode, nil) for *exec.ExitError
			// cases — unlike the original cmd.Wait, where non-zero exits
			// surfaced as err != nil. We have to check both here so a
			// crashed/killed ffmpeg (OOM, SIGKILL, corrupt input, etc.)
			// still gets marked failed instead of silently retaining
			// whatever status it had.
			if err != nil || exitCode != 0 {
				s.logger.Warn("Recording ffmpeg process exited unexpectedly",
					"id", id, "exitCode", exitCode, "error", err)
				active.recording.Status = "failed"
			}
		}
	}()

	// Startup health check: verify ffmpeg survives initial startup.
	// This catches wrong capture device, missing codec, unsupported format, etc.
	// FFmpeg typically fails within ~50ms when the capture device is unavailable;
	// 500ms is generous while still keeping the API responsive.
	startupCheck := time.NewTimer(500 * time.Millisecond)
	select {
	case err := <-done:
		startupCheck.Stop()
		// FFmpeg exited during startup — capture device likely unavailable
		if err != nil {
			return nil, fmt.Errorf("ffmpeg exited immediately during startup: %w", err)
		}
		return nil, fmt.Errorf("ffmpeg exited immediately during startup (exit code non-zero)")
	case <-startupCheck.C:
		// FFmpeg survived startup — recording is genuinely active
	}

	// Store active recording (only after startup health check passes)
	s.activeRecordings.Set(id, &activeRecording{
		recording: recording,
		cmd:       cmd,
		stdinPipe: stdinPipe,
		done:      done,
	})

	return recording, nil
}
