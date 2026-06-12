// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

// insertActive seeds the service with a live entry in the state StartRecording
// leaves it right after cmd.Start(): status "recording", done unsignalled.
func insertActive(s *RecordingService) (string, *activeRecording) {
	id := uuid.New().String()
	fileName := id + "_session_20260101_000000.mp4"
	active := &activeRecording{
		recording: &Recording{
			ID:        id,
			FileName:  fileName,
			FilePath:  filepath.Join(s.recordingsDir, fileName),
			StartTime: time.Now(),
			Status:    "recording",
		},
		done: make(chan error, 1),
	}
	s.activeRecordings.Set(id, active)
	return id, active
}

func TestConfirmStartupFailsWhenFFmpegDiesInWindow(t *testing.T) {
	s := newTestService(t)
	id, active := insertActive(s)

	// Simulate the wait goroutine observing an unexpected exit during the
	// startup window: markFailed first, then the done send (same ordering as
	// the real goroutine).
	active.markFailed(time.Now())
	active.done <- errors.New("exit status 1")

	err := s.confirmStartup(id, active, time.Second)
	if err == nil {
		t.Fatal("confirmStartup: expected error for ffmpeg death during startup window")
	}

	// The recording was never acknowledged, so it must not linger as a
	// phantom entry: list/stop/get all report not-found.
	if _, exists := s.activeRecordings.Get(id); exists {
		t.Fatal("entry still in active map after startup failure")
	}
	if _, err := s.StopRecording(id); !errors.Is(err, ErrRecordingNotFound) {
		t.Fatalf("StopRecording after startup failure: err = %v, want ErrRecordingNotFound", err)
	}
}

func TestConfirmStartupFailsOnCleanInstantExit(t *testing.T) {
	s := newTestService(t)
	id, active := insertActive(s)

	// Simulate the wait goroutine observing a clean instant exit: it pops the
	// entry itself and signals done with nil (Linux Reap also reports
	// non-zero exits as nil error). A recording whose process exited —
	// however cleanly — never started.
	s.activeRecordings.Pop(id)
	active.done <- nil

	err := s.confirmStartup(id, active, time.Second)
	if err == nil {
		t.Fatal("confirmStartup: expected error for instant clean exit")
	}
	if _, exists := s.activeRecordings.Get(id); exists {
		t.Fatal("entry still in active map after instant clean exit")
	}
}

func TestConfirmStartupSucceedsThenLaterFailureStillVisible(t *testing.T) {
	s := newTestService(t)
	id, active := insertActive(s)

	// ffmpeg survives the window: confirmStartup must return nil without
	// consuming done.
	if err := s.confirmStartup(id, active, 10*time.Millisecond); err != nil {
		t.Fatalf("confirmStartup with live process: unexpected error %v", err)
	}
	if _, exists := s.activeRecordings.Get(id); !exists {
		t.Fatal("entry vanished from active map after successful startup")
	}

	// ffmpeg dies after the successful start: the entry must remain visible
	// as failed, and StopRecording must not deadlock on the done channel
	// (its failed-snapshot path returns before waiting on done).
	active.markFailed(time.Now())
	active.done <- errors.New("exit status 1")

	rec, err := s.StopRecording(id)
	if err != nil {
		t.Fatalf("StopRecording after post-start failure: %v", err)
	}
	if rec.Status != "failed" {
		t.Fatalf("StopRecording after post-start failure: status = %q, want %q", rec.Status, "failed")
	}
	if _, exists := s.activeRecordings.Get(id); !exists {
		t.Fatal("failed entry vanished after StopRecording")
	}
}
