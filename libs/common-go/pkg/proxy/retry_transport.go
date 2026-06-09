// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const maxRetryBodySize = 256 << 10 // max body size buffered for retry

// retryTransport retries once on stale-connection errors (EOF/RST) and on
// multipart responses with empty bodies (stale conn returned headers only).
type retryTransport struct {
	base http.RoundTripper
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	bufferSmallBody(req)

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		if canReplay(req) && isStaleConnError(err) {
			slog.Warn("proxy: retrying after stale connection error",
				"method", req.Method, "path", req.URL.Path, "error", err)
			rewindBody(req)
			return t.base.RoundTrip(req)
		}
		return nil, err
	}

	if canReplay(req) && isMultipart(resp) {
		var peek [1]byte
		n, readErr := resp.Body.Read(peek[:])
		if n == 0 && readErr != nil {
			slog.Warn("proxy: retrying after empty multipart body",
				"method", req.Method, "path", req.URL.Path)
			resp.Body.Close()
			rewindBody(req)
			return t.base.RoundTrip(req)
		}
		if n > 0 {
			resp.Body = &prefixedBody{prefix: peek[:n], rest: resp.Body}
		}
	}

	return resp, nil
}

// bufferSmallBody captures small request bodies for replay; large uploads are skipped.
func bufferSmallBody(req *http.Request) {
	if req.GetBody != nil || req.Body == nil {
		return
	}
	if req.ContentLength < 0 || req.ContentLength > int64(maxRetryBodySize) {
		return
	}
	data, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		req.Body = io.NopCloser(bytes.NewReader(nil))
		return
	}
	req.ContentLength = int64(len(data))
	req.Body = io.NopCloser(bytes.NewReader(data))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
}

func canReplay(req *http.Request) bool { return req.GetBody != nil }

func rewindBody(req *http.Request) {
	if req.GetBody != nil {
		req.Body, _ = req.GetBody()
	}
}

func isStaleConnError(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	s := err.Error()
	return strings.Contains(s, "connection reset by peer") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "use of closed network connection")
}

func isMultipart(resp *http.Response) bool {
	return strings.HasPrefix(resp.Header.Get("Content-Type"), "multipart/")
}

// prefixedBody prepends already-read bytes back onto a ReadCloser.
type prefixedBody struct {
	prefix []byte
	off    int
	rest   io.ReadCloser
}

func (r *prefixedBody) Read(p []byte) (int, error) {
	if r.off < len(r.prefix) {
		n := copy(p, r.prefix[r.off:])
		r.off += n
		return n, nil
	}
	return r.rest.Read(p)
}

func (r *prefixedBody) Close() error { return r.rest.Close() }
