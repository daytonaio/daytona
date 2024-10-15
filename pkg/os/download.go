// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package os

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFile(ctx context.Context, url string, filename string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = os.MkdirAll(filepath.Dir(filename), os.ModePerm)
	if err != nil {
		return err
	}

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
		if ctx.Err() != nil {
			os.Remove(filename)
		}
	}()

	done := make(chan error, 1)

	go func() {
		_, err := io.Copy(out, resp.Body)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			os.Remove(filename)
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
