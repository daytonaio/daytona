// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateLabel(t *testing.T) {
	tests := []struct {
		name    string
		label   string
		wantErr bool
	}{
		{name: "valid simple", label: "my-recording", wantErr: false},
		{name: "valid with underscores and dots", label: "rec_1.0-test", wantErr: false},
		{name: "valid with spaces", label: "screen test 1", wantErr: false},
		{name: "empty string", label: "", wantErr: true},
		{name: "whitespace only", label: "   ", wantErr: true},
		{name: "leading dot", label: ".hidden", wantErr: true},
		{name: "forward slash traversal", label: "foo/bar", wantErr: true},
		{name: "backslash traversal", label: "foo\\bar", wantErr: true},
		{name: "special characters", label: "rec;rm -rf", wantErr: true},
		{name: "too long", label: strings.Repeat("a", 101), wantErr: true},
		{name: "max length exactly 100", label: strings.Repeat("a", 100), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLabel(tt.label)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLabel(%q) error = %v, wantErr %v", tt.label, err, tt.wantErr)
			}
		})
	}
}

func TestRecordingService_Basics(t *testing.T) {
	tempDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewRecordingService(logger, tempDir)

	if got := svc.GetRecordingsDir(); got != tempDir {
		t.Errorf("GetRecordingsDir() = %v, want %v", got, tempDir)
	}

	// Non-existent recording should return ErrRecordingNotFound
	_, err := svc.GetRecording("non-existent-id")
	if !errors.Is(err, ErrRecordingNotFound) {
		t.Errorf("GetRecording() error = %v, want ErrRecordingNotFound", err)
	}

	// Empty recordings dir should return empty list
	recordings, err := svc.ListRecordings()
	if err != nil {
		t.Fatalf("ListRecordings() unexpected error: %v", err)
	}
	if len(recordings) != 0 {
		t.Errorf("ListRecordings() returned %d items, want 0", len(recordings))
	}

	// Create a dummy mp4 file on disk with UUID in name
	dummyUUID := "123e4567-e89b-12d3-a456-426614174000"
	dummyFile := filepath.Join(tempDir, dummyUUID+"_test_20260916_120000.mp4")
	if err := os.WriteFile(dummyFile, []byte("fake mp4 content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should now find the completed recording on disk
	recordings, err = svc.ListRecordings()
	if err != nil {
		t.Fatalf("ListRecordings() error: %v", err)
	}
	if len(recordings) != 1 {
		t.Fatalf("ListRecordings() returned %d items, want 1", len(recordings))
	}
	if recordings[0].ID != dummyUUID {
		t.Errorf("recording ID = %q, want %q", recordings[0].ID, dummyUUID)
	}
	if recordings[0].Status != "completed" {
		t.Errorf("recording Status = %q, want %q", recordings[0].Status, "completed")
	}

	// GetRecording by ID
	rec, err := svc.GetRecording(dummyUUID)
	if err != nil {
		t.Fatalf("GetRecording(%q) error: %v", dummyUUID, err)
	}
	if rec.ID != dummyUUID {
		t.Errorf("rec.ID = %q, want %q", rec.ID, dummyUUID)
	}

	// DeleteRecording
	if err := svc.DeleteRecording(dummyUUID); err != nil {
		t.Fatalf("DeleteRecording(%q) error: %v", dummyUUID, err)
	}
	// File should be gone
	if _, err := os.Stat(dummyFile); !os.IsNotExist(err) {
		t.Errorf("file still exists after DeleteRecording")
	}
}

func TestStartRecording_InvalidLabel(t *testing.T) {
	tempDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewRecordingService(logger, tempDir)

	badLabel := "../traversal"
	_, err := svc.StartRecording(&badLabel)
	if !errors.Is(err, ErrInvalidLabel) {
		t.Errorf("StartRecording with bad label error = %v, want ErrInvalidLabel", err)
	}
}
