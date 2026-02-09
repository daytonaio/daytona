// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

// GetRecording returns a recording by ID (active or from filesystem)
func (s *RecordingService) GetRecording(id string) (*Recording, error) {
	// First check active recordings
	if active, exists := s.activeRecordings.Get(id); exists {
		recording := *active.recording
		return &recording, nil
	}

	// Search in completed recordings on disk
	recordings, err := s.ListRecordings()
	if err != nil {
		return nil, err
	}

	for _, rec := range recordings {
		if rec.ID == id {
			return &rec, nil
		}
	}

	return nil, ErrRecordingNotFound
}
