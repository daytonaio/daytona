// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	cmap "github.com/orcaman/concurrent-map/v2"
)

// RecordingService manages screen recording sessions
type RecordingService struct {
	activeRecordings cmap.ConcurrentMap[string, *activeRecording]
	recordingsDir    string
}

func NewRecordingService(recordingsDir string) *RecordingService {
	return &RecordingService{
		activeRecordings: cmap.New[*activeRecording](),
		recordingsDir:    recordingsDir,
	}
}

func (s *RecordingService) GetRecordingsDir() string {
	return s.recordingsDir
}
