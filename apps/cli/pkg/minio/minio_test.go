// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package minio

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// fixtureDirWithSymlink builds a directory that snapshot create would pack as
// Docker context: a regular file and a symlink whose target is larger than the
// link path. archive/tar reports "write too long" when a header Size does not
// match the bytes written (symlink / size mismatch, or the output tar itself
// being walked while it grows).
func fixtureDirWithSymlink(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("hello world this is longer than the link name"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target.txt", filepath.Join(dir, "link.txt")); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestWriteDirectoryToTar_SymlinkSizeMismatch(t *testing.T) {
	dir := fixtureDirWithSymlink(t)

	// Simulate `daytona snapshot create --dockerfile ... -c .`: the context tar
	// is created inside the walked tree and grows as entries are written.
	tarPath := filepath.Join(dir, CONTEXT_TAR_FILE_NAME)
	tarFile, err := os.Create(tarPath)
	if err != nil {
		t.Fatal(err)
	}
	defer tarFile.Close()

	tw := tar.NewWriter(tarFile)
	err = writeDirectoryToTar(tw, dir)
	if err != nil {
		t.Fatalf("packing context with symlink: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("closing tar writer: %v", err)
	}
	if _, err := tarFile.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	tr := tar.NewReader(tarFile)
	var sawTarget, sawLink bool
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		base := filepath.Base(hdr.Name)
		switch base {
		case "target.txt":
			sawTarget = true
			if hdr.Typeflag != tar.TypeReg {
				t.Fatalf("target.txt type = %q, want regular file", string(hdr.Typeflag))
			}
			body, err := io.ReadAll(tr)
			if err != nil {
				t.Fatal(err)
			}
			if string(body) != "hello world this is longer than the link name" {
				t.Fatalf("target.txt body = %q", body)
			}
		case "link.txt":
			sawLink = true
			if hdr.Typeflag != tar.TypeSymlink {
				t.Fatalf("link.txt type = %q, want symlink", string(hdr.Typeflag))
			}
			if hdr.Linkname != "target.txt" {
				t.Fatalf("link.txt Linkname = %q, want %q", hdr.Linkname, "target.txt")
			}
			if hdr.Size != 0 {
				t.Fatalf("link.txt Size = %d, want 0", hdr.Size)
			}
		case CONTEXT_TAR_FILE_NAME:
			t.Fatal("context tar must not be included in the archive")
		}
	}
	if !sawTarget || !sawLink {
		t.Fatalf("archive missing entries: target=%v link=%v", sawTarget, sawLink)
	}
}
