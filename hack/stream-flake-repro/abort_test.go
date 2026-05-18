// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

// Hostile-upstream test: the upstream finishes writing the multipart body but
// closes the underlying TCP socket *before* net/http's chunked-encoder flush
// can run. This is precisely what the CI logs showed ("readfrom tcp ...:
// unexpected EOF" on the proxy side). The point is to demonstrate that the
// pre-fix code path (defer mw.Close, no flush) is *vulnerable* to the same
// truncation pattern: any "connection closes mid-trailer" event drops the
// closing multipart boundary, and the SDK has no way to detect it.
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// upstreamMode controls how the synthetic upstream finishes a response.
type upstreamMode int

const (
	modePreFixAbort  upstreamMode = iota // defer mw.Close + hijack + close — simulates CI race
	modePostFixFlush                     // explicit mw.Close + explicit Flush before close — patched daemon
)

func runHostileServer(t *testing.T, mode upstreamMode) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go serveOnce(conn, mode)
		}
	}()
	return ln
}

func serveOnce(conn net.Conn, mode upstreamMode) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	br := bufio.NewReader(conn)
	if _, err := http.ReadRequest(br); err != nil {
		return
	}
	const boundary = "DAYTONA-FILE-BOUNDARY"
	bw := bufio.NewWriter(conn)
	// Write status line and headers manually so we control exactly what
	// reaches the wire — and what doesn't.
	_, _ = bw.WriteString("HTTP/1.1 200 OK\r\n")
	_, _ = bw.WriteString("Content-Type: multipart/form-data; boundary=" + boundary + "\r\n")
	_, _ = bw.WriteString("Transfer-Encoding: chunked\r\n")
	_, _ = bw.WriteString("\r\n")
	_ = bw.Flush()

	unit := []byte("progress-check-8993686a88ec4238b758d71cd6077b01")
	const perFile = 24 * 1024
	perFileAligned := (perFile / len(unit)) * len(unit)

	// Write the file part using a real multipart.Writer, but to a buffer we
	// chunk-encode ourselves so we can decide exactly when (and whether) the
	// closing boundary reaches the wire.
	writeChunk := func(p []byte) {
		_, _ = fmt.Fprintf(bw, "%x\r\n", len(p))
		_, _ = bw.Write(p)
		_, _ = bw.WriteString("\r\n")
		_ = bw.Flush()
	}

	// Opening boundary + part headers.
	hdr := "--" + boundary + "\r\n"
	hdr += "Content-Type: application/octet-stream\r\n"
	hdr += `Content-Disposition: form-data; name="file"; filename="f.bin"` + "\r\n"
	hdr += "Content-Length: " + strconv.Itoa(perFileAligned) + "\r\n"
	hdr += "\r\n"
	writeChunk([]byte(hdr))

	// Payload, in 4KB chunks like the daemon does.
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
		writeChunk(buf[:n])
		written += n
	}

	// Pre-fix path: write closing boundary into the bufio writer but DO NOT
	// flush. Then close the socket abruptly, mimicking "handler returns,
	// connection closes before chunked-encoder/bufio flush completes".
	closing := "\r\n--" + boundary + "--\r\n"
	if mode == modePreFixAbort {
		// Stage the closing boundary in the bufio buffer without flushing.
		_, _ = fmt.Fprintf(bw, "%x\r\n", len(closing))
		_, _ = bw.WriteString(closing)
		_, _ = bw.WriteString("\r\n")
		// Final chunked terminator — also unflushed.
		_, _ = bw.WriteString("0\r\n\r\n")
		// Abort: close the TCP connection with the buffer still in memory.
		if tcp, ok := conn.(*net.TCPConn); ok {
			_ = tcp.SetLinger(0) // RST instead of FIN — drops buffered data.
		}
		return // deferred conn.Close() will RST.
	}

	// Post-fix path: write closing boundary, flush, then close cleanly.
	writeChunk([]byte(closing))
	_, _ = bw.WriteString("0\r\n\r\n")
	_ = bw.Flush()
}

func TestHostileUpstream_PreFixDropsClosingBoundary(t *testing.T) {
	t.Parallel()
	ln := runHostileServer(t, modePreFixAbort)
	defer ln.Close()

	const (
		concurrency = 32
		iterations  = 50
	)

	var truncations atomic.Int64
	var wg sync.WaitGroup
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}
			for i := 0; i < iterations; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
					"http://"+ln.Addr().String()+"/x", nil)
				resp, err := client.Do(req)
				cancel()
				if err != nil {
					truncations.Add(1)
					continue
				}
				mt, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
				if err != nil || mt != "multipart/form-data" {
					_ = resp.Body.Close()
					truncations.Add(1)
					continue
				}
				mr := multipart.NewReader(resp.Body, params["boundary"])
				part, err := mr.NextPart()
				if err != nil {
					_ = resp.Body.Close()
					truncations.Add(1)
					continue
				}
				_, _ = io.ReadAll(part)
				_, eofErr := mr.NextPart()
				_ = resp.Body.Close()
				if eofErr != io.EOF {
					// This is the silent-truncation pattern: parser observed
					// an error / non-EOF instead of a clean message end.
					truncations.Add(1)
				}
			}
		}()
	}
	wg.Wait()

	total := int64(concurrency * iterations)
	t.Logf("pre-fix hostile upstream: %d/%d requests observed parser-level truncation",
		truncations.Load(), total)
	if truncations.Load() == 0 {
		t.Fatal("expected at least some truncations from RST-after-stage; got 0 — reproducer is broken")
	}
}


