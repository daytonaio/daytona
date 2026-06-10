// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

var (
	ErrRecordingNotFound    = errors.New("recording not found")
	ErrRecordingNotActive   = errors.New("recording is not active")
	ErrRecordingStillActive = errors.New("cannot delete an active recording")
	ErrFFmpegNotFound       = errors.New("ffmpeg not found in PATH")
	ErrInvalidLabel         = errors.New("invalid label: must be 1-100 characters, cannot be blank, cannot start with a dot, cannot contain consecutive dots (..), and can only contain ASCII letters, numbers, spaces, dots, underscores, and hyphens")
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

// activeRecording holds the state of a currently running recording.
//
// mu guards recording field mutations: the wait goroutine (markFailed) and
// StopRecording both mutate the shared *Recording while List/Get may read it
// concurrently via IterBuffered, so all reads go through snapshot().
type activeRecording struct {
	mu        sync.Mutex
	recording *Recording
	cmd       *exec.Cmd
	stdinPipe io.WriteCloser
	done      chan error
}

// snapshot returns a copy of the recording taken under the lock.
func (a *activeRecording) snapshot() Recording {
	a.mu.Lock()
	defer a.mu.Unlock()
	return *a.recording
}

// markFailed transitions the recording to "failed" after an unexpected ffmpeg
// exit, recording end time, duration, and the size of any partial file. It
// also closes the stdin pipe so the dead process's pipe handle is released.
// Callers must invoke it before signalling the done channel so that
// StopRecording observes the failure after receiving from done.
func (a *activeRecording) markFailed(at time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.recording.Status = "failed"
	a.recording.EndTime = &at
	duration := at.Sub(a.recording.StartTime).Seconds()
	a.recording.DurationSeconds = &duration
	if fileInfo, err := os.Stat(a.recording.FilePath); err == nil {
		size := fileInfo.Size()
		a.recording.SizeBytes = &size
	}
	if a.stdinPipe != nil {
		_ = a.stdinPipe.Close()
	}
}
