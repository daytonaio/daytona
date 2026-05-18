// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build server

// Hostile upstream: writes a multipart/form-data body followed by only the
// PREFIX of the closing boundary, then closes the connection cleanly (FIN).
// The python-multipart parser keeps those prefix bytes in its boundary
// look-ahead buffer waiting for the rest, then silently discards them at
// finalize() — exactly the pattern observed in CI ("got 24023 bytes,
// want 24064" with no error raised).
package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go serveBad(conn)
	}
}

func serveBad(c net.Conn) {
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(10 * time.Second))
	br := bufio.NewReader(c)
	if _, err := http.ReadRequest(br); err != nil {
		return
	}
	const boundary = "DAYTONA-FILE-BOUNDARY"
	bw := bufio.NewWriter(c)
	_, _ = bw.WriteString("HTTP/1.1 200 OK\r\n")
	_, _ = bw.WriteString("Content-Type: multipart/form-data; boundary=" + boundary + "\r\n")
	_, _ = bw.WriteString("Transfer-Encoding: chunked\r\n")
	_, _ = bw.WriteString("\r\n")
	_ = bw.Flush()

	unit := []byte("progress-check-8993686a88ec4238b758d71cd6077b01")
	const perFile = 24 * 1024
	perFileAligned := (perFile / len(unit)) * len(unit)

	writeChunk := func(p []byte) {
		_, _ = fmt.Fprintf(bw, "%x\r\n", len(p))
		_, _ = bw.Write(p)
		_, _ = bw.WriteString("\r\n")
		_ = bw.Flush()
	}

	hdr := "--" + boundary + "\r\n" +
		"Content-Type: application/octet-stream\r\n" +
		`Content-Disposition: form-data; name="file"; filename="f.bin"` + "\r\n" +
		"Content-Length: " + strconv.Itoa(perFileAligned) + "\r\n\r\n"
	writeChunk([]byte(hdr))

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

	// Send only the prefix of the closing boundary, then end the stream
	// cleanly. To the python-multipart parser this is indistinguishable
	// from "a partial boundary that might still complete" — it buffers
	// the prefix and drops it at finalize().
	partial := []byte("\r\n--" + boundary)
	writeChunk(partial)
	_, _ = bw.WriteString("0\r\n\r\n") // chunked terminator
	_ = bw.Flush()
	// Clean FIN.
}
