// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package os

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func DownloadFile(c *gin.Context, url string, filename string) error {
	var ctx context.Context
	if c != nil {
		ctx = c.Request.Context()
	} else {
		ctx = context.Background()
	}

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
