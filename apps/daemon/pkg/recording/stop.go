// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"os"
	"time"
)

// StopRecording stops an active recording session
func (s *RecordingService) StopRecording(id string) (*Recording, error) {
	active, exists := s.activeRecordings.Pop(id)
	if !exists {
		return nil, ErrRecordingNotFound
	}

	// Send 'q' to ffmpeg stdin for graceful shutdown
	// This allows ffmpeg to properly finalize the video file
	if active.stdinPipe != nil {
		_, err := active.stdinPipe.Write([]byte("q"))
		if err != nil {
			s.logger.Warn("Failed to send quit signal to ffmpeg", "error", err)
		}
		active.stdinPipe.Close()
	}

	// Wait for ffmpeg to finish by waiting on the done channel
	select {
	case <-active.done:
		// Process exited normally
	case <-time.After(10 * time.Second):
		// Force kill if it doesn't exit gracefully
		s.logger.Warn("Recording did not stop gracefully, force killing", "id", id)
		if active.cmd.Process != nil {
			err := active.cmd.Process.Kill()
			if err != nil {
				s.logger.Error("Failed to force kill recording", "id", id, "error", err)
			}
		}
		// Still wait for the done channel to avoid goroutine leak
		<-active.done
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

	s.logger.Debug("Stopped recording", "id", id, "durationSeconds", duration)

	return active.recording, nil
}
