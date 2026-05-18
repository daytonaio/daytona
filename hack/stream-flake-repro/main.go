// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build !server

// Reproduces the streaming-multipart truncation flake observed in CI:
//   - upstream: gin handler that writes a multipart/form-data body using the
//     *current* download_files.go pattern (defer mw.Close, no explicit flush).
//   - proxy:    common-go NewProxyRequestHandler (configurable FlushInterval).
//   - client:   N concurrent HTTP downloaders that assert (a) byte count
//               matches, (b) the body ends with the closing boundary, and
//               (c) the multipart parser finishes with io.EOF.
//
// Run with `-fixed=false` to model the original pre-fix behaviour (defer Close,
// no explicit Flush, proxy FlushInterval=0). Run with `-fixed=true` to verify
// the patched behaviour. With `-races` running concurrently in goroutines and
// proxy + upstream chained through real TCP loopback, we get realistic enough
// scheduling to trigger the same race CI hits.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type result struct {
	id       int
	ok       bool
	reason   string
	gotBytes int
}

func main() {
	var (
		fixed       = flag.Bool("fixed", false, "use the patched (explicit Close+Flush) handler and FlushInterval=-1 on the proxy")
		concurrency = flag.Int("c", 64, "concurrent downloaders")
		iterations  = flag.Int("n", 200, "downloads per worker")
		payloadKB   = flag.Int("kb", 32, "payload size per file in KB")
		files       = flag.Int("files", 1, "files per response")
	)
	flag.Parse()

	payloadUnit := []byte("progress-check-8993686a88ec4238b758d71cd6077b01") // 47 bytes
	perFile := *payloadKB * 1024
	totalPerFile := (perFile / len(payloadUnit)) * len(payloadUnit)

	upstream := httptest.NewServer(streamingHandler(*fixed, payloadUnit, totalPerFile, *files))
	defer upstream.Close()

	proxy := newProxy(upstream.URL, *fixed)
	proxyServer := httptest.NewServer(proxy)
	defer proxyServer.Close()

	log.Printf("upstream: %s", upstream.URL)
	log.Printf("proxy:    %s  (fixed=%v)", proxyServer.URL, *fixed)
	log.Printf("payload:  %d bytes per file, %d files per response", totalPerFile, *files)
	log.Printf("load:     %d workers x %d iterations = %d total downloads",
		*concurrency, *iterations, *concurrency*(*iterations))

	results := make(chan result, *concurrency*(*iterations))
	start := time.Now()

	var wg sync.WaitGroup
	for w := 0; w < *concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			client := &http.Client{Timeout: 30 * time.Second}
			for i := 0; i < *iterations; i++ {
				r := downloadOne(client, proxyServer.URL, payloadUnit, totalPerFile, *files)
				r.id = workerID*(*iterations) + i
				results <- r
			}
		}(w)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var ok, bad int64
	failures := []result{}
	for r := range results {
		if r.ok {
			atomic.AddInt64(&ok, 1)
		} else {
			atomic.AddInt64(&bad, 1)
			if len(failures) < 5 {
				failures = append(failures, r)
			}
		}
	}

	elapsed := time.Since(start)
	total := ok + bad
	log.Printf("done: %d ok, %d bad, %.2f%% bad, total %d in %s (%.0f req/s)",
		ok, bad, float64(bad)*100/float64(total), total, elapsed, float64(total)/elapsed.Seconds())
	for _, f := range failures {
		log.Printf("  FAIL[%d]: %s (got %d bytes)", f.id, f.reason, f.gotBytes)
	}
	if bad > 0 {
		os.Exit(1)
	}
}

func downloadOne(client *http.Client, proxyURL string, unit []byte, perFile, fileCount int) result {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, proxyURL+"/download", nil)
	resp, err := client.Do(req)
	if err != nil {
		return result{reason: fmt.Sprintf("client.Do: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return result{reason: fmt.Sprintf("status %d", resp.StatusCode)}
	}

	ct := resp.Header.Get("Content-Type")
	mt, params, err := mime.ParseMediaType(ct)
	if err != nil || mt != "multipart/form-data" {
		return result{reason: fmt.Sprintf("bad content-type %q", ct)}
	}
	boundary := params["boundary"]
	if boundary == "" {
		return result{reason: "missing boundary"}
	}

	mr := multipart.NewReader(resp.Body, boundary)
	parts := 0
	totalBytes := 0
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result{reason: fmt.Sprintf("NextPart: %v (after %d parts, %d bytes)", err, parts, totalBytes), gotBytes: totalBytes}
		}

		expectedSize := perFile
		if cl := part.Header.Get("Content-Length"); cl != "" {
			expectedSize, _ = strconv.Atoi(cl)
		}

		buf, err := io.ReadAll(part)
		_ = part.Close()
		totalBytes += len(buf)
		if err != nil {
			return result{reason: fmt.Sprintf("read part %d: %v", parts, err), gotBytes: totalBytes}
		}
		if len(buf) != expectedSize {
			return result{reason: fmt.Sprintf("part %d short: got %d want %d", parts, len(buf), expectedSize), gotBytes: totalBytes}
		}
		// Verify payload integrity (catch silent bit-flips / wrong-offset reads)
		for i := 0; i < len(buf); i++ {
			if buf[i] != unit[i%len(unit)] {
				return result{reason: fmt.Sprintf("part %d corrupt at byte %d", parts, i), gotBytes: totalBytes}
			}
		}
		parts++
	}

	if parts != fileCount {
		return result{reason: fmt.Sprintf("part count: got %d want %d", parts, fileCount), gotBytes: totalBytes}
	}
	return result{ok: true, gotBytes: totalBytes}
}

// streamingHandler mirrors the daemon's download_files.go behaviour. When
// `fixed` is false it reproduces the original pre-fix behaviour: defer
// mw.Close, no explicit Flush. When true it does explicit Close + Flush.
func streamingHandler(fixed bool, payloadUnit []byte, perFile, fileCount int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const boundary = "DAYTONA-FILE-BOUNDARY"
		w.Header().Set("Content-Type", "multipart/form-data; boundary="+boundary)
		w.WriteHeader(http.StatusOK)

		mw := multipart.NewWriter(w)
		_ = mw.SetBoundary(boundary)
		if !fixed {
			defer mw.Close() // original pattern
		}

		for f := 0; f < fileCount; f++ {
			hdr := textproto.MIMEHeader{}
			hdr.Set("Content-Type", "application/octet-stream")
			hdr.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="f%d.bin"`, f))
			hdr.Set("Content-Length", strconv.Itoa(perFile))
			part, err := mw.CreatePart(hdr)
			if err != nil {
				return
			}
			// Write the payload in 4KB chunks to mimic io.Copy from a file.
			// Each byte at file offset i must be payloadUnit[i % len(unit)] so
			// the client can verify integrity from a single global pattern.
			written := 0
			buf := make([]byte, 4096)
			for written < perFile {
				n := len(buf)
				if written+n > perFile {
					n = perFile - written
				}
				for j := 0; j < n; j++ {
					buf[j] = payloadUnit[(written+j)%len(payloadUnit)]
				}
				m, err := part.Write(buf[:n])
				if err != nil {
					return
				}
				written += m
			}
			if fixed {
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}

		if fixed {
			if err := mw.Close(); err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

func newProxy(upstreamURL string, fixed bool) http.Handler {
	u, _ := url.Parse(upstreamURL)
	rp := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.Host = u.Host
		},
	}
	if fixed {
		rp.FlushInterval = -1
	}
	return rp
}
