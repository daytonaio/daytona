// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// replayableReq builds a request whose GetBody is populated (as net/http does
// for in-memory bodies), so canReplay reports true and the multipart-retry
// path is exercised.
func replayableReq(t *testing.T, method, url, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	return req
}

// TestRetryTransport_MultipartReturnsConnToPool verifies that forwarding a
// small multipart response fully drains and releases the upstream connection,
// so subsequent requests reuse it instead of dialing fresh — the regression
// for the heap leak where each forwarded request orphaned its connection.
func TestRetryTransport_MultipartReturnsConnToPool(t *testing.T) {
	body := strings.Repeat("x", 1024)
	var newConns int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "multipart/form-data; boundary=abc")
		io.WriteString(w, body)
	}))
	srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			atomic.AddInt64(&newConns, 1)
		}
	}
	defer srv.Close()

	rt := &retryTransport{base: &http.Transport{MaxIdleConns: 10, MaxIdleConnsPerHost: 10}}
	for i := 0; i < 5; i++ {
		resp, err := rt.RoundTrip(replayableReq(t, http.MethodPost, srv.URL, "hello"))
		if err != nil {
			t.Fatalf("RoundTrip: %v", err)
		}
		got, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if string(got) != body {
			t.Fatalf("body mismatch: got %d bytes, want %d", len(got), len(body))
		}
	}
	if c := atomic.LoadInt64(&newConns); c != 1 {
		t.Errorf("expected connection reuse (1 new conn), got %d", c)
	}
}

// TestRetryTransport_LargeMultipartDrainsOnClose verifies the over-cap
// streaming path: a partial read followed by Close must drain the remainder so
// the connection still returns to the pool. Before the fix, prefixedBody.Close
// closed without draining, discarding the connection on every request.
func TestRetryTransport_LargeMultipartDrainsOnClose(t *testing.T) {
	big := strings.Repeat("y", maxRetryBodySize*2)
	var newConns int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "multipart/mixed; boundary=z")
		io.WriteString(w, big)
	}))
	srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			atomic.AddInt64(&newConns, 1)
		}
	}
	defer srv.Close()

	rt := &retryTransport{base: &http.Transport{MaxIdleConns: 10, MaxIdleConnsPerHost: 10}}
	for i := 0; i < 3; i++ {
		resp, err := rt.RoundTrip(replayableReq(t, http.MethodPost, srv.URL, "x"))
		if err != nil {
			t.Fatalf("RoundTrip: %v", err)
		}
		var p [16]byte
		resp.Body.Read(p[:]) // partial read only
		resp.Body.Close()    // Close must drain the rest
	}
	if c := atomic.LoadInt64(&newConns); c != 1 {
		t.Errorf("expected drain-on-close to reuse conn (1 new conn), got %d", c)
	}
}

// TestRetryTransport_TruncatedMultipartSurfacesError verifies that a body cut
// off mid-read is not silently served from the buffer as a complete response:
// the client must receive the buffered prefix followed by a read error, not a
// clean EOF on truncated data.
func TestRetryTransport_TruncatedMultipartSurfacesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Errorf("Hijack: %v", err)
			return
		}
		// Promise 100 bytes but deliver 7, then drop the connection.
		io.WriteString(conn, "HTTP/1.1 200 OK\r\n"+
			"Content-Type: multipart/form-data; boundary=abc\r\n"+
			"Content-Length: 100\r\n\r\npartial")
		conn.Close()
	}))
	defer srv.Close()

	rt := &retryTransport{base: &http.Transport{MaxIdleConns: 10, MaxIdleConnsPerHost: 10}}
	resp, err := rt.RoundTrip(replayableReq(t, http.MethodPost, srv.URL, "x"))
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	got, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr == nil {
		t.Fatalf("expected read error for truncated body, got clean EOF with %d bytes", len(got))
	}
	if string(got) != "partial" {
		t.Errorf("buffered prefix: got %q, want %q", got, "partial")
	}
}

// TestRetryTransport_EmptyMultipartRetries verifies that an empty multipart
// body (stale conn returned headers only) triggers a single replay and the
// caller receives the second, non-empty response.
func TestRetryTransport_EmptyMultipartRetries(t *testing.T) {
	var calls int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "multipart/form-data; boundary=abc")
		if atomic.AddInt64(&calls, 1) == 1 {
			return // empty body on first attempt
		}
		io.WriteString(w, "second-try-body")
	}))
	defer srv.Close()

	rt := &retryTransport{base: &http.Transport{MaxIdleConns: 10, MaxIdleConnsPerHost: 10}}
	resp, err := rt.RoundTrip(replayableReq(t, http.MethodPost, srv.URL, "x"))
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	got, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(got) != "second-try-body" {
		t.Fatalf("got %q, want %q", got, "second-try-body")
	}
	if c := atomic.LoadInt64(&calls); c != 2 {
		t.Fatalf("expected 2 upstream calls (retry), got %d", c)
	}
}
