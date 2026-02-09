// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
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
			log.Warnf("Failed to send quit signal to ffmpeg: %v", err)
		}
		active.stdinPipe.Close()
	}

	// Wait for ffmpeg to finish by waiting on the done channel
	select {
	case <-active.done:
		// Process exited normally
	case <-time.After(10 * time.Second):
		// Force kill if it doesn't exit gracefully
		log.Warnf("Recording %s did not stop gracefully, force killing", id)
		if active.cmd.Process != nil {
			err := active.cmd.Process.Kill()
			if err != nil {
				log.Errorf("Failed to force kill recording %s: %v", id, err)
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

	log.Debugf("Stopped recording %s, duration: %.2f seconds", id, duration)

	return active.recording, nil
}
