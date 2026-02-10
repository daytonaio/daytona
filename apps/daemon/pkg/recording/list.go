// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ListRecordings returns all recordings (active and completed)
func (s *RecordingService) ListRecordings() ([]Recording, error) {
	recordings := []Recording{}

	// Add active recordings
	for item := range s.activeRecordings.IterBuffered() {
		recordings = append(recordings, *item.Val.recording)
	}

	// Scan recordings directory for completed recordings
	if _, err := os.Stat(s.recordingsDir); os.IsNotExist(err) {
		return recordings, nil
	}

	entries, err := os.ReadDir(s.recordingsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read recordings directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only include MP4 files
		if filepath.Ext(entry.Name()) != ".mp4" {
			continue
		}

		// Skip files that are currently being recorded
		isActive := false
		for item := range s.activeRecordings.IterBuffered() {
			if item.Val.recording.FileName == entry.Name() {
				isActive = true
				break
			}
		}

		if isActive {
			continue
		}

		filePath := filepath.Join(s.recordingsDir, entry.Name())
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		// Extract ID from filename (format: {id}_{label}_{timestamp}.mp4 or {id}_session_{timestamp}.mp4)
		// The ID is the first part before the underscore (UUID format)
		fileName := entry.Name()
		var recordingID string
		if idx := strings.Index(fileName, "_"); idx > 0 {
			potentialID := fileName[:idx]
			// Validate it's a UUID
			if _, err := uuid.Parse(potentialID); err == nil {
				recordingID = potentialID
			}
		}
		// Fallback to generating ID from file path for legacy recordings without ID in filename
		if recordingID == "" {
			recordingID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(filePath)).String()
		}

		// Create recording entry from file info
		// Use file modification time as a proxy for end time
		modTime := fileInfo.ModTime()
		size := fileInfo.Size()

		recording := Recording{
			ID:        recordingID,
			FileName:  fileName,
			FilePath:  filePath,
			StartTime: modTime, // Approximation - actual start time unknown for old recordings
			EndTime:   &modTime,
			Status:    "completed",
			SizeBytes: &size,
		}

		recordings = append(recordings, recording)
	}

	return recordings, nil
}
