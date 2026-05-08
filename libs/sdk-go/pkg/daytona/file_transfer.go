// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

type downloadStreamCloser struct {
	partReader io.Reader
	response   *http.Response
}

type progressReader struct {
	inner      io.Reader
	onProgress func(DownloadProgress)
	total      int64
	expected   int64
}

func (d *downloadStreamCloser) Read(p []byte) (int, error) {
	return d.partReader.Read(p)
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.inner.Read(buf)
	if n > 0 {
		p.total += int64(n)
		p.onProgress(DownloadProgress{
			BytesReceived: p.total,
			TotalBytes:    p.expected,
		})
	}
	return n, err
}

func (d *downloadStreamCloser) Close() error {
	if d.response == nil || d.response.Body == nil {
		return nil
	}
	return d.response.Body.Close()
}

func streamDownloadFile(cfg *toolbox.Configuration, remotePath string, ctx context.Context, onProgress func(DownloadProgress)) (io.ReadCloser, error) {
	if len(cfg.Servers) == 0 {
		return nil, errors.NewDaytonaError("Toolbox client is not configured", 0, nil)
	}

	requestBody, err := json.Marshal(toolbox.NewFilesDownloadRequest([]string{remotePath}))
	if err != nil {
		return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to encode download request: %v", err), 0, nil)
	}

	endpoint := strings.TrimRight(cfg.Servers[0].URL, "/") + "/files/bulk-download"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to create download request: %v", err), 0, nil)
	}

	for key, value := range cfg.DefaultHeader {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "multipart/form-data")

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to download file stream: %v", err), 0, nil)
	}

	if resp.StatusCode >= http.StatusMultipleChoices {
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to read download error response: %v", readErr), resp.StatusCode, resp.Header)
		}

		return nil, errors.NewDaytonaErrorFromBody(body, resp.StatusCode, resp.Header)
	}

	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		_ = resp.Body.Close()
		return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to parse multipart response: %v", err), resp.StatusCode, resp.Header)
	}

	boundary := params["boundary"]
	if boundary == "" {
		_ = resp.Body.Close()
		return nil, errors.NewDaytonaError("Missing multipart boundary in download response", resp.StatusCode, resp.Header)
	}

	reader := multipart.NewReader(resp.Body, boundary)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			_ = resp.Body.Close()
			return nil, errors.NewDaytonaError("File stream not found in download response", resp.StatusCode, resp.Header)
		}
		if err != nil {
			_ = resp.Body.Close()
			return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to read multipart download response: %v", err), resp.StatusCode, resp.Header)
		}

		switch part.FormName() {
		case "file":
			expected := int64(0)
			if cl := part.Header.Get("Content-Length"); cl != "" {
				if n, parseErr := strconv.ParseInt(cl, 10, 64); parseErr == nil {
					expected = n
				}
			}

			var reader io.Reader = part
			if onProgress != nil {
				reader = &progressReader{inner: part, onProgress: onProgress, expected: expected}
			}
			return &downloadStreamCloser{partReader: reader, response: resp}, nil
		case "error":
			body, readErr := io.ReadAll(part)
			_ = resp.Body.Close()
			if readErr != nil {
				return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to read download error part: %v", readErr), resp.StatusCode, resp.Header)
			}
			return nil, errors.NewDaytonaErrorFromBody(body, resp.StatusCode, resp.Header)
		}
	}
}

type uploadProgressReader struct {
	inner      io.Reader
	onProgress func(UploadProgress)
	total      int64
}

func (p *uploadProgressReader) Read(buf []byte) (int, error) {
	n, err := p.inner.Read(buf)
	if n > 0 {
		p.total += int64(n)
		if p.onProgress != nil {
			p.onProgress(UploadProgress{BytesSent: p.total})
		}
	}
	return n, err
}

// streamUploadFile builds a multipart/form-data request body via io.Pipe and feeds the
// caller's io.Reader directly through it, so neither the SDK nor net/http buffer the
// payload. Context cancellation aborts both the multipart writer goroutine and the
// in-flight HTTP request.
func streamUploadFile(cfg *toolbox.Configuration, remotePath string, source io.Reader, opts *uploadStreamConfig, ctx context.Context) error {
	if len(cfg.Servers) == 0 {
		return errors.NewDaytonaError("Toolbox client is not configured", 0, nil)
	}

	wrapped := io.Reader(source)
	if opts.onProgress != nil {
		wrapped = &uploadProgressReader{inner: source, onProgress: opts.onProgress}
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		writeErr := func() error {
			if err := mw.WriteField("files[0].path", remotePath); err != nil {
				return err
			}
			part, err := mw.CreateFormFile("files[0].file", remotePath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(part, wrapped); err != nil {
				return err
			}
			if err := mw.Close(); err != nil {
				return err
			}
			return nil
		}()

		_ = pw.CloseWithError(writeErr)
		errCh <- writeErr
	}()

	endpoint := strings.TrimRight(cfg.Servers[0].URL, "/") + "/files/bulk-upload"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, pr)
	if err != nil {
		_ = pw.CloseWithError(err)
		<-errCh
		return errors.NewDaytonaError(fmt.Sprintf("Failed to create upload request: %v", err), 0, nil)
	}
	for key, value := range cfg.DefaultHeader {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		_ = pw.CloseWithError(err)
		<-errCh
		return errors.NewDaytonaError(fmt.Sprintf("Failed to upload file: %v", err), 0, nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		returnErr := errors.NewDaytonaErrorFromBody(body, resp.StatusCode, resp.Header)
		_ = pw.CloseWithError(returnErr)
		<-errCh
		return returnErr
	}

	if writeErr := <-errCh; writeErr != nil {
		return writeErr
	}

	return nil
}
