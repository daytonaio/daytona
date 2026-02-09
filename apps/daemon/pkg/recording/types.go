// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"io"
	"os/exec"
	"time"
)

var (
	ErrRecordingNotFound    = errors.New("recording not found")
	ErrRecordingNotActive   = errors.New("recording is not active")
	ErrRecordingStillActive = errors.New("cannot delete an active recording")
	ErrFFmpegNotFound       = errors.New("ffmpeg not found in PATH")
	ErrInvalidLabel         = errors.New("invalid label: must be 1-100 characters, cannot start with dot, cannot contain path separators (/ or \\), and can only contain letters, numbers, spaces, dots, underscores, and hyphens")
)

type Recording struct {
	ID              string
	FileName        string
	FilePath        string
	StartTime       time.Time
	EndTime         *time.Time
	Status          string
	DurationSeconds *float64
	SizeBytes       *int64
}

// activeRecording holds the state of a currently running recording
type activeRecording struct {
	recording *Recording
	cmd       *exec.Cmd
	stdinPipe io.WriteCloser
	done      chan error
}
