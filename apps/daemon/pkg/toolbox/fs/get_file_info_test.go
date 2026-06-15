// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"os"
	"path/filepath"
	"testing"
)

// Regression test for the Windows drift where the per-platform getFileInfo
// copy omitted ModifiedAt and serialized the zero time on the wire. Runs on
// every platform the package builds for.
func TestGetFileInfoSetsModifiedAt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	info, err := getFileInfo(path)
	if err != nil {
		t.Fatalf("getFileInfo: %v", err)
	}

	if info.ModifiedAt.IsZero() {
		t.Fatal("ModifiedAt is zero — must be populated on every platform")
	}
	if info.ModTime != info.ModifiedAt.String() {
		t.Fatalf("ModTime %q and ModifiedAt %q must derive from the same timestamp", info.ModTime, info.ModifiedAt)
	}
}
