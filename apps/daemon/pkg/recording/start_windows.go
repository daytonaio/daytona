//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
)

func (s *RecordingService) StartRecording(label *string) (*Recording, error) {
	if err := os.MkdirAll(s.recordingsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create recordings directory: %w", err)
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, ErrFFmpegNotFound
	}

	id := uuid.New().String()
	now := time.Now()
	timestamp := now.Format("20060102_150405")

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

	recording := &Recording{
		ID:        id,
		FileName:  fileName,
		FilePath:  filePath,
		StartTime: now,
		Status:    "recording",
	}

	cmd := exec.Command(ffmpegPath,
		"-f", "gdigrab",
		"-framerate", "30",
		"-i", "desktop",
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-y",
		filePath,
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	s.logger.Debug("Started recording", "id", id, "path", filePath)

	done := make(chan error, 1)

	s.activeRecordings.Set(id, &activeRecording{
		recording: recording,
		cmd:       cmd,
		stdinPipe: stdinPipe,
		done:      done,
	})

	go func() {
		err := cmd.Wait()
		done <- err

		if active, exists := s.activeRecordings.Pop(id); exists {
			if err != nil {
				s.logger.Warn("Recording ffmpeg process exited unexpectedly", "id", id, "error", err)
				active.recording.Status = "failed"
			}
		}
	}()

	return recording, nil
}
