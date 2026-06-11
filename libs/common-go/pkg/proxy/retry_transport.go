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
		// Buffer up to maxRetryBodySize+1 instead of peek-reading the live
		// body: peeking a single byte and re-wrapping leaves net/http's body
		// half-read, so the persistConn is never drained and returned to the
		// idle pool. Reading one byte past the cap lets us tell whether the
		// whole body fits in the buffer.
		buf, readErr := io.ReadAll(io.LimitReader(resp.Body, maxRetryBodySize+1))
		if readErr != nil {
			// The body broke mid-read. Serving buf as-is would hand the client
			// a truncated body with a clean EOF, so either replay (nothing read
			// yet, stale conn) or surface the prefix followed by the error.
			resp.Body.Close()
			if len(buf) == 0 && isStaleConnError(readErr) {
				slog.Warn("proxy: retrying after body read error",
					"method", req.Method, "path", req.URL.Path, "error", readErr)
				rewindBody(req)
				return t.base.RoundTrip(req)
			}
			resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(buf), errReader{readErr}))
			return resp, nil
		}
		if len(buf) == 0 {
			// Empty multipart body (stale conn returned headers only): drain
			// and close so the connection returns to the pool, then replay.
			drainClose(resp.Body)
			slog.Warn("proxy: retrying after empty multipart body",
				"method", req.Method, "path", req.URL.Path)
			rewindBody(req)
			return t.base.RoundTrip(req)
		}
		if len(buf) <= maxRetryBodySize {
			// Whole body fit in the buffer, so the upstream body is at EOF;
			// release its connection to the pool and serve from memory.
			drainClose(resp.Body)
			resp.Body = io.NopCloser(bytes.NewReader(buf))
		} else {
			// Body exceeds the buffer: non-empty for certain, so stream the
			// remainder. prefixedBody.Close drains the rest on completion.
			resp.Body = &prefixedBody{prefix: buf, rest: resp.Body}
		}
	}

	return resp, nil
}

// drainClose reads any unread bytes from body and closes it, so the underlying
// connection is returned to the idle pool instead of being discarded.
func drainClose(body io.ReadCloser) {
	io.Copy(io.Discard, body)
	body.Close()
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

// errReader yields err on every read, so a buffered prefix can be followed by
// the original mid-body failure instead of a clean EOF.
type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

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

// Close drains any unread remainder before closing so the underlying
// connection can return to the idle pool rather than being discarded.
func (r *prefixedBody) Close() error {
	io.Copy(io.Discard, r.rest)
	return r.rest.Close()
}
