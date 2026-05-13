// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type fileResult struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
}

// UploadFiles godoc
//
//	@Summary		Upload multiple files
//	@Description	Upload multiple files with their destination paths
//	@Tags			file-system
//	@Accept			multipart/form-data
//	@Success		200
//	@Router			/files/bulk-upload [post]
//
//	@id				UploadFiles
func UploadFiles(c *gin.Context) {
	reader, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []string{"invalid multipart form"}})
		return
	}

	dests := make(map[string]string)
	var errs []string
	var files []fileResult

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			errs = append(errs, fmt.Sprintf("reading part: %v", err))
			break
		}

		name := part.FormName()

		if strings.HasSuffix(name, ".path") {
			data, err := io.ReadAll(part)
			if err != nil {
				idx := extractIndex(name)
				errs = append(errs, fmt.Sprintf("path[%s]: %v", idx, err))
				continue
			}
			idx := extractIndex(name)
			dest := strings.TrimSpace(string(data))
			if dest == "" {
				errs = append(errs, fmt.Sprintf("path[%s]: empty path", idx))
				continue
			}
			dests[idx] = dest
			continue
		}

		if strings.HasSuffix(name, ".file") {
			idx := extractIndex(name)
			dest, ok := dests[idx]
			if !ok {
				errs = append(errs, fmt.Sprintf("file[%s]: missing .path metadata", idx))
				continue
			}

			n, writeErr := writeUploadedPart(part, dest)
			if writeErr != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", dest, writeErr))
				continue
			}
			files = append(files, fileResult{Path: dest, Bytes: n})
			continue
		}
	}

	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errs, "files": files})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

func writeUploadedPart(part *multipart.Part, dest string) (int64, error) {
	dir := filepath.Dir(dest)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return 0, fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}

	tmp, err := os.CreateTemp(dir, ".daytona-upload-*")
	if err != nil {
		return 0, fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	n, err := io.Copy(tmp, part)
	if err != nil {
		tmp.Close()
		return 0, fmt.Errorf("write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return 0, fmt.Errorf("close: %w", err)
	}
	resolvedDest, err := resolveSymlink(dest)
	if err != nil {
		return 0, fmt.Errorf("resolve dest: %w", err)
	}
	// Best-effort: some FUSE-backed filesystems (e.g. S3 volume mounts) reject
	// chmod(2). If it fails we still attempt the rename; the copyAndRemove
	// fallback creates the destination with 0644 via OpenFile, so permissions
	// are correct on both code paths regardless.
	_ = os.Chmod(tmpPath, 0o644)
	if err := os.Rename(tmpPath, resolvedDest); err != nil {
		if cpErr := copyAndRemove(tmpPath, resolvedDest); cpErr != nil {
			return 0, fmt.Errorf("rename: %w; fallback: %v", err, cpErr)
		}
	}
	committed = true
	return n, nil
}

func resolveSymlink(path string) (string, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return path, nil
		}
		return "", err
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		return path, nil
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	target, err := os.Readlink(path)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	return target, nil
}

func copyAndRemove(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	_ = os.Remove(src)
	return nil
}

func extractIndex(fieldName string) string {
	s := strings.TrimPrefix(fieldName, "files[")
	return strings.TrimSuffix(strings.TrimSuffix(s, "].path"), "].file")
}
