// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
)

// createTarFromFile creates a tar archive containing a single file
func createTarFromFile(filePath, destName string) (io.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	header := &tar.Header{
		Name: destName,
		Mode: 0755,
		Size: fileInfo.Size(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return nil, err
	}

	if _, err := io.Copy(tw, file); err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}
