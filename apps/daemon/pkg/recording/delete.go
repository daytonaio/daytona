// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"fmt"
	"os"
)

// DeleteRecording deletes a recording by ID
func (s *RecordingService) DeleteRecording(id string) error {
	// Check if it's an in-memory recording. A failed entry may exist only in
	// memory (ffmpeg died before producing a playable file), so it must be
	// removable even when no mp4 exists; live recordings must be stopped first.
	if active, exists := s.activeRecordings.Get(id); exists {
		if active.snapshot().Status != "failed" {
			return ErrRecordingStillActive
		}

		s.activeRecordings.Remove(id)
		if err := os.Remove(active.recording.FilePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete recording file: %w", err)
		}

		s.logger.Debug("Deleted failed recording", "id", id, "filePath", active.recording.FilePath)
		return nil
	}

	// Find the recording
	recording, err := s.GetRecording(id)
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

	s.logger.Debug("Deleted recording", "id", id, "filePath", recording.FilePath)

	return nil
}
