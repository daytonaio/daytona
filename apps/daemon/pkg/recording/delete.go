// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"fmt"
	"os"
)

// DeleteRecording deletes a recording by ID
func (s *RecordingService) DeleteRecording(id string) error {
	// Check if it's an active recording
	if _, exists := s.activeRecordings.Get(id); exists {
		return ErrRecordingStillActive
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
