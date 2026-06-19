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

	// ffmpeg already exited unexpectedly: report the failure instead of
	// fabricating "completed". Re-insert the entry (Pop claimed it) so the
	// failed recording stays visible until DeleteRecording removes it. Skip
	// the quit/wait dance: the done value was already consumed or would be
	// consumed exactly once by the stop that raced the failure.
	if rec := active.snapshot(); rec.Status == "failed" {
		s.activeRecordings.Set(id, active)
		return &rec, nil
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

	// Re-check: the process may have died on its own while we waited. The
	// wait goroutine marks failed before signalling done, so the snapshot is
	// authoritative here.
	if rec := active.snapshot(); rec.Status == "failed" {
		s.activeRecordings.Set(id, active)
		return &rec, nil
	}

	// Update recording metadata under the lock; a concurrent ListRecordings
	// may still be reading this entry through a pointer it captured earlier.
	now := time.Now()
	active.mu.Lock()
	active.recording.EndTime = &now
	active.recording.Status = "completed"

	duration := now.Sub(active.recording.StartTime).Seconds()
	active.recording.DurationSeconds = &duration

	// Get file size
	if fileInfo, err := os.Stat(active.recording.FilePath); err == nil {
		size := fileInfo.Size()
		active.recording.SizeBytes = &size
	}
	rec := *active.recording
	active.mu.Unlock()

	s.logger.Debug("Stopped recording", "id", id, "durationSeconds", duration)

	return &rec, nil
}
