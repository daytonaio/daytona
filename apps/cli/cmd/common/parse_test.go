// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
)

func TestParseKeyValuePairs(t *testing.T) {
	tests := []struct {
		name    string
		entries []string
		want    map[string]string
		wantErr bool
	}{
		{name: "empty input", entries: nil, want: map[string]string{}},
		{name: "single pair", entries: []string{"FOO=bar"}, want: map[string]string{"FOO": "bar"}},
		{name: "multiple pairs", entries: []string{"A=1", "B=2"}, want: map[string]string{"A": "1", "B": "2"}},
		{name: "value contains equals", entries: []string{"URL=http://x?a=b"}, want: map[string]string{"URL": "http://x?a=b"}},
		{name: "empty value allowed", entries: []string{"FOO="}, want: map[string]string{"FOO": ""}},
		{name: "last entry wins for duplicate key", entries: []string{"A=1", "A=2"}, want: map[string]string{"A": "2"}},
		{name: "missing separator", entries: []string{"FOO"}, wantErr: true},
		{name: "empty key", entries: []string{"=bar"}, wantErr: true},
		{name: "empty entry", entries: []string{""}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeyValuePairs(tt.entries, "env")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseKeyValuePairs(%v) expected error, got nil", tt.entries)
				}
				var cliErr *clierr.Error
				if !errors.As(err, &cliErr) || cliErr.Category != clierr.CategoryUsage {
					t.Errorf("ParseKeyValuePairs(%v) error = %v, want usage-category clierr", tt.entries, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseKeyValuePairs(%v) unexpected error: %v", tt.entries, err)
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("ParseKeyValuePairs(%v) = %v, want %v", tt.entries, got, tt.want)
			}
		})
	}
}

func TestParseVolumeSpecs(t *testing.T) {
	tests := []struct {
		name      string
		entries   []string
		wantId    string
		wantMount string
		wantErr   bool
	}{
		{name: "valid spec", entries: []string{"vol-123:/data"}, wantId: "vol-123", wantMount: "/data"},
		{name: "mount path contains colon", entries: []string{"vol-123:/a:b"}, wantId: "vol-123", wantMount: "/a:b"},
		{name: "missing separator", entries: []string{"vol-123"}, wantErr: true},
		{name: "empty volume id", entries: []string{":/data"}, wantErr: true},
		{name: "empty mount path", entries: []string{"vol-123:"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVolumeSpecs(tt.entries)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseVolumeSpecs(%v) expected error, got nil", tt.entries)
				}
				var cliErr *clierr.Error
				if !errors.As(err, &cliErr) || cliErr.Category != clierr.CategoryUsage {
					t.Errorf("ParseVolumeSpecs(%v) error = %v, want usage-category clierr", tt.entries, err)
				}
				if !strings.Contains(err.Error(), "expected VOLUME:MOUNT_PATH") {
					t.Errorf("ParseVolumeSpecs(%v) error = %q, want it to mention expected VOLUME:MOUNT_PATH", tt.entries, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseVolumeSpecs(%v) unexpected error: %v", tt.entries, err)
			}
			if len(got) != 1 {
				t.Fatalf("ParseVolumeSpecs(%v) returned %d volumes, want 1", tt.entries, len(got))
			}
			if got[0].VolumeId != tt.wantId || got[0].MountPath != tt.wantMount {
				t.Errorf("ParseVolumeSpecs(%v) = %s:%s, want %s:%s", tt.entries, got[0].VolumeId, got[0].MountPath, tt.wantId, tt.wantMount)
			}
		})
	}
}

func TestReadKeyValueFile(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		if _, err := ReadKeyValueFile(filepath.Join(t.TempDir(), "nope.env")); err == nil {
			t.Fatal("ReadKeyValueFile expected error for missing file, got nil")
		}
	})

	t.Run("valid file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.env")
		if err := os.WriteFile(path, []byte("FOO=bar\nBAZ=qux\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := ReadKeyValueFile(path)
		if err != nil {
			t.Fatalf("ReadKeyValueFile unexpected error: %v", err)
		}
		want := map[string]string{"FOO": "bar", "BAZ": "qux"}
		if !maps.Equal(got, want) {
			t.Errorf("ReadKeyValueFile = %v, want %v", got, want)
		}
	})
}

func TestResolveKeyValuePairs(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "values.env")
	if err := os.WriteFile(filePath, []byte("A=file\nB=file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		entries  []string
		filePath string
		want     map[string]string
		wantErr  bool
	}{
		{name: "cli only", entries: []string{"A=cli"}, want: map[string]string{"A": "cli"}},
		{name: "file only", filePath: filePath, want: map[string]string{"A": "file", "B": "file"}},
		{name: "cli overrides file per key", entries: []string{"B=cli", "C=cli"}, filePath: filePath, want: map[string]string{"A": "file", "B": "cli", "C": "cli"}},
		{name: "missing file", entries: []string{"A=cli"}, filePath: filepath.Join(dir, "nope.env"), wantErr: true},
		{name: "invalid cli entry", entries: []string{"broken"}, filePath: filePath, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveKeyValuePairs(tt.entries, tt.filePath, "env")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ResolveKeyValuePairs(%v, %q) expected error, got nil", tt.entries, tt.filePath)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveKeyValuePairs(%v, %q) unexpected error: %v", tt.entries, tt.filePath, err)
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("ResolveKeyValuePairs(%v, %q) = %v, want %v", tt.entries, tt.filePath, got, tt.want)
			}
		})
	}
}
