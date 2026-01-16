// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"time"
)

// RecordingStatus represents the current state of a recording
type RecordingStatus string

const (
	StatusRecording RecordingStatus = "recording"
	StatusCompleted RecordingStatus = "completed"
	StatusFailed    RecordingStatus = "failed"
)

// StartRecordingRequest represents the request to start a new recording
type StartRecordingRequest struct {
	Label string `json:"label,omitempty"` // Optional custom label for the recording
} // @name StartRecordingRequest

// StopRecordingRequest represents the request to stop an active recording
type StopRecordingRequest struct {
	ID string `json:"id" validate:"required"` // Recording ID to stop
} // @name StopRecordingRequest

// Recording represents a recording session (active or completed)
type Recording struct {
	ID              string          `json:"id"`
	FileName        string          `json:"fileName"`
	FilePath        string          `json:"filePath"`
	StartTime       time.Time       `json:"startTime"`
	EndTime         *time.Time      `json:"endTime,omitempty"`
	Status          RecordingStatus `json:"status"`
	DurationSeconds *float64        `json:"durationSeconds,omitempty"`
	SizeBytes       *int64          `json:"sizeBytes,omitempty"`
} // @name Recording

// StartRecordingResponse represents the response after starting a recording
type StartRecordingResponse struct {
	ID        string    `json:"id"`
	FileName  string    `json:"fileName"`
	FilePath  string    `json:"filePath"`
	StartTime time.Time `json:"startTime"`
	Status    string    `json:"status"`
} // @name StartRecordingResponse

// StopRecordingResponse represents the response after stopping a recording
type StopRecordingResponse struct {
	ID              string  `json:"id"`
	FilePath        string  `json:"filePath"`
	DurationSeconds float64 `json:"durationSeconds"`
	Status          string  `json:"status"`
} // @name StopRecordingResponse

// ListRecordingsResponse represents the response containing all recordings
type ListRecordingsResponse struct {
	Recordings []Recording `json:"recordings"`
} // @name ListRecordingsResponse

// GetRecordingResponse represents the response for a single recording
type GetRecordingResponse struct {
	Recording
} // @name GetRecordingResponse
