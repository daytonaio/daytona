// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

// This test simulates the actual CI failure pattern: a streaming upstream
// whose response handler returns immediately after writing, racing the kernel
// socket flush against connection close. Run with:
//
//	GOWORK=off go test -count=1 -run Race -timeout=60s ./...
//
// On enough iterations under enough goroutines, the multipart parser observes
// short reads if the daemon (a) doesn't explicitly flush before returning and
// (b) the network layer truncates the trailing bytes.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRace_UpstreamTruncatesViaConnectionClose(t *testing.T) {
	t.Parallel()
	// Knobs: tune up if you want to hunt harder on a busy machine.
	const (
		concurrency = 64
		iterations  = 200
		perFile     = 24 * 1024
	)

	unit := []byte("progress-check-8993686a88ec4238b758d71cd6077b01")
	perFileAligned := (perFile / len(unit)) * len(unit)

	// Pre-fix upstream behaviour: defer mw.Close(), no explicit Flush.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const boundary = "DAYTONA-FILE-BOUNDARY"
		w.Header().Set("Content-Type", "multipart/form-data; boundary="+boundary)
		w.WriteHeader(http.StatusOK)

		mw := multipart.NewWriter(w)
		_ = mw.SetBoundary(boundary)
		defer mw.Close() // <-- the original pattern

		hdr := textproto.MIMEHeader{}
		hdr.Set("Content-Type", "application/octet-stream")
		hdr.Set("Content-Disposition", `form-data; name="file"; filename="f.bin"`)
		hdr.Set("Content-Length", strconv.Itoa(perFileAligned))
		part, err := mw.CreatePart(hdr)
		if err != nil {
			return
		}
		buf := make([]byte, 4096)
		written := 0
		for written < perFileAligned {
			n := len(buf)
			if written+n > perFileAligned {
				n = perFileAligned - written
			}
			for j := 0; j < n; j++ {
				buf[j] = unit[(written+j)%len(unit)]
			}
			m, err := part.Write(buf[:n])
			if err != nil {
				return
			}
			written += m
		}
	}))
	defer upstream.Close()

	// Default ReverseProxy: FlushInterval=0 — matches pre-fix behaviour.
	u, _ := url.Parse(upstream.URL)
	proxy := httptest.NewServer(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.Host = u.Host
		},
	})
	defer proxy.Close()

	verifyOnce := func(t *testing.T, client *http.Client) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, proxy.URL+"/x", nil)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("do: %w", err)
		}
		defer resp.Body.Close()
		mt, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil || mt != "multipart/form-data" {
			return fmt.Errorf("ct: %v", err)
		}
		mr := multipart.NewReader(resp.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			return fmt.Errorf("nextpart: %w", err)
		}
		body, err := io.ReadAll(part)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		if len(body) != perFileAligned {
			return fmt.Errorf("short: got %d want %d", len(body), perFileAligned)
		}
		if _, err := mr.NextPart(); !errors.Is(err, io.EOF) {
			return fmt.Errorf("trailing parts or err: %v", err)
		}
		return nil
	}

	var fails atomic.Int64
	var wg sync.WaitGroup
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 10 * time.Second}
			for i := 0; i < iterations; i++ {
				if err := verifyOnce(t, client); err != nil {
					fails.Add(1)
					if fails.Load() <= 3 {
						t.Logf("failure %d: %v", fails.Load(), err)
					}
				}
			}
		}()
	}
	wg.Wait()
	if fails.Load() > 0 {
		t.Logf("Observed %d failures out of %d requests — reproducer is hot.", fails.Load(), int64(concurrency*iterations))
	} else {
		t.Logf("No failures in %d requests — race window did not open on this run.", int64(concurrency*iterations))
	}
}
