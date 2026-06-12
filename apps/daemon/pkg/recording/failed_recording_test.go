// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestService(t *testing.T) *RecordingService {
	t.Helper()
	return NewRecordingService(slog.New(slog.NewTextHandler(io.Discard, nil)), t.TempDir())
}

// insertFailed seeds the service with an entry in the state the wait
// goroutine leaves behind after an unexpected ffmpeg exit: marked failed,
// done channel already signalled, no file on disk.
func insertFailed(s *RecordingService) string {
	id := uuid.New().String()
	fileName := id + "_session_20260101_000000.mp4"
	done := make(chan error, 1)
	active := &activeRecording{
		recording: &Recording{
			ID:        id,
			FileName:  fileName,
			FilePath:  filepath.Join(s.recordingsDir, fileName),
			StartTime: time.Now().Add(-time.Minute),
			Status:    "recording",
		},
		done: done,
	}
	s.activeRecordings.Set(id, active)
	active.markFailed(time.Now())
	done <- errors.New("exit status 1")
	return id
}

func TestStopFailedRecordingReturnsFailure(t *testing.T) {
	s := newTestService(t)
	id := insertFailed(s)

	rec, err := s.StopRecording(id)
	if err != nil {
		t.Fatalf("StopRecording on failed entry: unexpected error %v", err)
	}
	if rec.Status != "failed" {
		t.Fatalf("StopRecording on failed entry: status = %q, want %q", rec.Status, "failed")
	}
	if rec.EndTime == nil || rec.DurationSeconds == nil {
		t.Fatalf("StopRecording on failed entry: EndTime/DurationSeconds not set: %+v", rec)
	}

	// The failed entry must stay visible until deleted; a second stop must
	// not hang on the consumed done channel and must report the same state.
	if _, exists := s.activeRecordings.Get(id); !exists {
		t.Fatal("failed entry vanished from active map after StopRecording")
	}
	rec2, err := s.StopRecording(id)
	if err != nil {
		t.Fatalf("second StopRecording on failed entry: unexpected error %v", err)
	}
	if rec2.Status != "failed" {
		t.Fatalf("second StopRecording: status = %q, want %q", rec2.Status, "failed")
	}
}

func TestListAndGetShowFailedRecording(t *testing.T) {
	s := newTestService(t)
	id := insertFailed(s)

	recs, err := s.ListRecordings()
	if err != nil {
		t.Fatalf("ListRecordings: %v", err)
	}
	found := false
	for _, r := range recs {
		if r.ID == id {
			found = true
			if r.Status != "failed" {
				t.Fatalf("ListRecordings: status = %q, want %q", r.Status, "failed")
			}
		}
	}
	if !found {
		t.Fatal("failed recording missing from ListRecordings")
	}

	rec, err := s.GetRecording(id)
	if err != nil {
		t.Fatalf("GetRecording: %v", err)
	}
	if rec.Status != "failed" {
		t.Fatalf("GetRecording: status = %q, want %q", rec.Status, "failed")
	}
}

func TestDeleteFailedRecordingWithoutFile(t *testing.T) {
	s := newTestService(t)
	id := insertFailed(s)

	if err := s.DeleteRecording(id); err != nil {
		t.Fatalf("DeleteRecording on failed entry without mp4: %v", err)
	}
	if _, exists := s.activeRecordings.Get(id); exists {
		t.Fatal("failed entry still in active map after delete")
	}
	if _, err := s.GetRecording(id); !errors.Is(err, ErrRecordingNotFound) {
		t.Fatalf("GetRecording after delete: err = %v, want ErrRecordingNotFound", err)
	}
}

func TestDeleteFailedRecordingRemovesPartialFile(t *testing.T) {
	s := newTestService(t)
	if err := os.MkdirAll(s.recordingsDir, 0755); err != nil {
		t.Fatal(err)
	}
	id := insertFailed(s)
	active, _ := s.activeRecordings.Get(id)
	if err := os.WriteFile(active.recording.FilePath, []byte("partial"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteRecording(id); err != nil {
		t.Fatalf("DeleteRecording on failed entry with partial mp4: %v", err)
	}
	if _, err := os.Stat(active.recording.FilePath); !os.IsNotExist(err) {
		t.Fatalf("partial file still present after delete: %v", err)
	}
}

func TestDeleteLiveRecordingRejected(t *testing.T) {
	s := newTestService(t)
	id := uuid.New().String()
	s.activeRecordings.Set(id, &activeRecording{
		recording: &Recording{
			ID:        id,
			Status:    "recording",
			StartTime: time.Now(),
			FilePath:  filepath.Join(s.recordingsDir, id+".mp4"),
		},
		done: make(chan error, 1),
	})

	if err := s.DeleteRecording(id); !errors.Is(err, ErrRecordingStillActive) {
		t.Fatalf("DeleteRecording on live entry: err = %v, want ErrRecordingStillActive", err)
	}
}
