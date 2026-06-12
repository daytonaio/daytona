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

	"github.com/daytonaio/daemon/pkg/winsession"
	"github.com/google/uuid"
)

// consoleSessionTimeout is short: the daemon runs as a service in session 0,
// and gdigrab can only capture the interactive desktop, so by the time a
// recording is requested AutoLogon must already have produced the console
// session. If it hasn't, fail fast instead of stalling the API call.
const consoleSessionTimeout = 5 * time.Second

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

	// The daemon runs as SYSTEM in session 0, whose desktop has nothing to
	// capture; spawn ffmpeg with the interactive console user's token so
	// gdigrab sees the real desktop (same mechanism as the computer-use
	// plugin spawn).
	token, err := winsession.ActiveConsoleUserToken(consoleSessionTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve console session for recording: %w", err)
	}
	// CreateProcessAsUser references the token into the child during Start();
	// our duplicated handle stays ours and must outlive cmd.Start(), so close
	// it at function exit on every path.
	defer token.Close()

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Token:         syscall.Token(token),
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

	active := &activeRecording{
		recording: recording,
		cmd:       cmd,
		stdinPipe: stdinPipe,
		done:      done,
	}
	s.activeRecordings.Set(id, active)

	go func() {
		err := cmd.Wait()

		if err != nil {
			// Unexpected exit (a graceful 'q' stop yields exit code 0 and the
			// entry is already popped by StopRecording). Keep the entry
			// visible as "failed" so list/stop/delete can report and clean it
			// up instead of having it silently vanish. markFailed must happen
			// before the done send so StopRecording observes the failure.
			if active, exists := s.activeRecordings.Get(id); exists {
				s.logger.Warn("Recording ffmpeg process exited unexpectedly", "id", id, "error", err)
				active.markFailed(time.Now())
			}
		} else {
			// Clean exit outside StopRecording (e.g. external quit): drop the
			// active entry; the finalized file on disk represents it.
			s.activeRecordings.Pop(id)
		}

		done <- err
	}()

	return recording, nil
}
