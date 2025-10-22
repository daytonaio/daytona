// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"log/slog"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// RecordingService manages screen recording sessions
type RecordingService struct {
	logger           *slog.Logger
	activeRecordings cmap.ConcurrentMap[string, *activeRecording]
	recordingsDir    string
}

func NewRecordingService(logger *slog.Logger, recordingsDir string) *RecordingService {
	return &RecordingService{
		logger:           logger.With(slog.String("component", "recording_service")),
		activeRecordings: cmap.New[*activeRecording](),
		recordingsDir:    recordingsDir,
	}
}

func (s *RecordingService) GetRecordingsDir() string {
	return s.recordingsDir
}
