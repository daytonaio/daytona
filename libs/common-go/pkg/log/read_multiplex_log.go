// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"time"
)

// Stream prefixes for multiplexing stdout/stderr in logs
var (
	STDOUT_PREFIX = []byte{0x01, 0x01, 0x01}
	STDERR_PREFIX = []byte{0x02, 0x02, 0x02}
)

func ReadMultiplexedLog(ctx context.Context, logReader io.Reader, follow bool, stdoutChan chan []byte, stderrChan chan []byte, errChan chan error) {
	reader := bufio.NewReader(logReader)

	// Accumulated buffer across reads so we can safely detect prefixes
	var buf []byte

	const (
		streamNone = iota
		streamStdout
		streamStderr
	)

	currentStream := streamNone
	maxPrefixLen := len(STDOUT_PREFIX)
	if l := len(STDERR_PREFIX); l > maxPrefixLen {
		maxPrefixLen = l
	}

	// Helper to determine which prefix occurs next in buf and at what index.
	findNextPrefix := func(b []byte) (idx int, stream int, prefixLen int) {
		idxOut := bytes.Index(b, STDOUT_PREFIX)
		idxErr := bytes.Index(b, STDERR_PREFIX)

		// No prefix found at all.
		if idxOut == -1 && idxErr == -1 {
			return -1, streamNone, 0
		}

		// Only one of them found.
		if idxOut != -1 && (idxErr == -1 || idxOut <= idxErr) {
			return idxOut, streamStdout, len(STDOUT_PREFIX)
		}
		return idxErr, streamStderr, len(STDERR_PREFIX)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			chunk := make([]byte, 1024)
			n, err := reader.Read(chunk)

			// Append any bytes that were actually read.
			if n > 0 {
				buf = append(buf, chunk[:n]...)
			}

			atEOF := false
			if err != nil {
				if err == io.EOF {
					atEOF = true
				} else {
					// Report non-EOF errors but still try to process any accumulated data.
					select {
					case errChan <- err:
					case <-ctx.Done():
						return
					}
				}
			}

			// Process the buffer: route bytes between prefixes to the appropriate stream.
			for {
				if currentStream == streamNone {
					// We are not currently in any stream: look for the first prefix.
					idx, stream, prefixLen := findNextPrefix(buf)
					if idx == -1 {
						// No prefix yet. To avoid losing a prefix that may be split across reads,
						// keep at most (maxPrefixLen-1) trailing bytes; discard any earlier noise.
						if !atEOF && len(buf) > maxPrefixLen-1 {
							buf = buf[len(buf)-(maxPrefixLen-1):]
						}
						break
					}

					// Discard anything before the first recognized prefix.
					if idx > 0 {
						buf = buf[idx:]
						idx = 0
					}

					// Consume the prefix and set the current stream.
					buf = buf[prefixLen:]
					currentStream = stream
					continue
				}

				// We are currently in a stream: look for the next prefix.
				idx, nextStream, prefixLen := findNextPrefix(buf)
				if idx == -1 {
					// No next prefix in the buffer.
					if atEOF {
						// At EOF, flush whatever remains in the buffer to the current stream.
						if len(buf) > 0 {
							out := make([]byte, len(buf))
							copy(out, buf)
							if currentStream == streamStdout {
								select {
								case <-ctx.Done():
									return
								case stdoutChan <- out:
								}
							} else if currentStream == streamStderr {
								select {
								case <-ctx.Done():
									return
								case stderrChan <- out:
								}
							}
						}
						buf = nil
					}
					// In the non-EOF case, keep buf as-is so that a prefix split across reads
					// is not broken; we'll continue once more data arrives.
					break
				}

				// Data up to the next prefix belongs to the current stream.
				if idx > 0 {
					chunkData := make([]byte, idx)
					copy(chunkData, buf[:idx])
					if currentStream == streamStdout {
						select {
						case <-ctx.Done():
							return
						case stdoutChan <- chunkData:
						}
					} else if currentStream == streamStderr {
						select {
						case <-ctx.Done():
							return
						case stderrChan <- chunkData:
						}
					}
				}

				// Consume the data and the next prefix, then switch streams.
				buf = buf[idx+prefixLen:]
				currentStream = nextStream
			}

			if err != nil {
				if err == io.EOF {
					if !follow {
						// Signal EOF and stop if we are not following.
						errChan <- io.EOF
						return
					}
					// If following, just continue waiting for more data.
					select {
					case <-ctx.Done():
						return
					case <-time.After(50 * time.Millisecond):
					}
				} else {
					// Err already sent to channel above
					return
				}
			}
		}
	}
}
