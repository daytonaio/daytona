// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"time"

	"github.com/daytonaio/daemon/pkg/recording"
)

// Recording represents a recording session (active or completed)
type RecordingDTO struct {
	ID              string     `json:"id" validate:"required"`
	FileName        string     `json:"fileName" validate:"required"`
	FilePath        string     `json:"filePath" validate:"required"`
	StartTime       time.Time  `json:"startTime" validate:"required"`
	EndTime         *time.Time `json:"endTime,omitempty"`
	Status          string     `json:"status" validate:"required"`
	DurationSeconds *float64   `json:"durationSeconds,omitempty"`
	SizeBytes       *int64     `json:"sizeBytes,omitempty"`
} // @name Recording

// StartRecordingRequest represents the request to start a new recording
type StartRecordingRequest struct {
	Label *string `json:"label,omitempty"`
} // @name StartRecordingRequest

// StopRecordingRequest represents the request to stop an active recording
type StopRecordingRequest struct {
	ID string `json:"id" validate:"required"`
} // @name StopRecordingRequest

// ListRecordingsResponse represents the response containing all recordings
type ListRecordingsResponse struct {
	Recordings []RecordingDTO `json:"recordings" validate:"required"`
} // @name ListRecordingsResponse

func RecordingToDTO(r *recording.Recording) *RecordingDTO {
	return &RecordingDTO{
		ID:              r.ID,
		FileName:        r.FileName,
		FilePath:        r.FilePath,
		StartTime:       r.StartTime,
		EndTime:         r.EndTime,
		Status:          r.Status,
		DurationSeconds: r.DurationSeconds,
		SizeBytes:       r.SizeBytes,
	}
}
