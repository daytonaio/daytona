// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/gorilla/websocket"

	common_errors "github.com/daytonaio/common-go/pkg/errors"

	log "github.com/sirupsen/logrus"
)

type FetchLogsOptions struct {
	IsCombinedOutput   bool
	IsWebsocketUpgrade bool
	Follow             bool
}

func (s *SessionService) GetSessionCommandLogs(sessionId, commandId string, request *http.Request, responseWriter http.ResponseWriter, opts FetchLogsOptions) ([]byte, error) {
	session, ok := s.sessions[sessionId]
	if !ok {
		return nil, common_errors.NewNotFoundError(errors.New("session not found"))
	}

	command, ok := s.sessions[sessionId].commands[commandId]
	if !ok {
		return nil, common_errors.NewNotFoundError(errors.New("command not found"))
	}

	logFilePath, exitCodeFilePath := command.LogFilePath(session.Dir(s.configDir))

	if opts.IsWebsocketUpgrade {
		logFile, err := os.Open(logFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, common_errors.NewNotFoundError(err)
			}
			if os.IsPermission(err) {
				return nil, common_errors.NewForbiddenError(err)
			}
			return nil, common_errors.NewBadRequestError(err)
		}
		defer logFile.Close()
		ReadLog(request, responseWriter, opts.Follow, logFile, util.ReadLogWithExitCode, exitCodeFilePath, func(conn *websocket.Conn, messages chan []byte, errors chan error) {
			var buffer []byte
			for {
				select {
				case <-session.ctx.Done():
					// Flush any remaining bytes in buffer before closing
					if opts.IsCombinedOutput && len(buffer) > 0 {
						remainingData := flushRemainingBuffer(&buffer)
						if len(remainingData) > 0 {
							err := conn.WriteMessage(websocket.BinaryMessage, remainingData)
							if err != nil {
								log.Error(err)
							}
						}
					}
					err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
					if err != nil {
						log.Error(err)
					}
					conn.Close()
					return
				case msg := <-messages:
					if opts.IsCombinedOutput {
						// Process chunks with buffering to handle prefixes split across chunks
						processedData := processLogChunkWithPrefixFiltering(msg, &buffer)
						if len(processedData) > 0 {
							err := conn.WriteMessage(websocket.BinaryMessage, processedData)
							if err != nil {
								errors <- err
								return
							}
						}
					} else {
						err := conn.WriteMessage(websocket.BinaryMessage, msg)
						if err != nil {
							errors <- err
							return
						}
					}
				case <-errors:
					// Stream ended, flush any remaining bytes in buffer
					if opts.IsCombinedOutput && len(buffer) > 0 {
						remainingData := flushRemainingBuffer(&buffer)
						if len(remainingData) > 0 {
							writeErr := conn.WriteMessage(websocket.BinaryMessage, remainingData)
							if writeErr != nil {
								log.Error(writeErr)
							}
						}
					}
					// The error will be handled by the main ReadLog function
					return
				}
			}
		})
		return nil, nil
	}

	logBytes, err := os.ReadFile(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, common_errors.NewNotFoundError(err)
		}
		if os.IsPermission(err) {
			return nil, common_errors.NewForbiddenError(err)
		}
		return nil, common_errors.NewBadRequestError(err)
	}

	if opts.IsCombinedOutput {
		// remove prefixes from log bytes
		logBytes = bytes.ReplaceAll(bytes.ReplaceAll(logBytes, STDOUT_PREFIX, []byte{}), STDERR_PREFIX, []byte{})
	}

	return logBytes, nil
}

// ReadLog reads from the logReader and writes to the websocket.
// TLogData is the type of the message to be read from the logReader
func ReadLog[TLogData any](request *http.Request, responseWriter http.ResponseWriter, follow bool, logReader io.Reader, readFunc func(context.Context, io.Reader, bool, string, chan TLogData, chan error), exitCodeFilePath string, wsWriteFunc func(*websocket.Conn, chan TLogData, chan error)) {
	ws, err := upgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		log.Error(err)
		return
	}

	defer func() {
		closeErr := websocket.CloseNormalClosure
		if !errors.Is(err, io.EOF) {
			closeErr = websocket.CloseInternalServerErr
		}
		err := ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(closeErr, ""), time.Now().Add(time.Second))
		if err != nil {
			log.Trace(err)
		}
		ws.Close()
	}()

	msgChannel := make(chan TLogData)
	errChannel := make(chan error)
	ctx, cancel := context.WithCancel(request.Context())

	defer cancel()
	go readFunc(ctx, logReader, follow, exitCodeFilePath, msgChannel, errChannel)
	go wsWriteFunc(ws, msgChannel, errChannel)

	readErr := make(chan error)
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			readErr <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case err = <-errChannel:
			if err != nil {
				if !errors.Is(err, io.EOF) {
					log.Error(err)
				}
				cancel()
				return
			}
		case err := <-readErr:
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				log.Error(err)
			}
			if err != nil {
				return
			}
		}
	}
}

// processLogChunkWithPrefixFiltering processes log chunks with buffering to handle prefixes split across chunks
func processLogChunkWithPrefixFiltering(chunk []byte, buffer *[]byte) []byte {
	// Append new chunk to buffer
	*buffer = append(*buffer, chunk...)

	var result []byte
	processed := 0

	for processed < len(*buffer) {
		// Check if we have enough bytes to check for prefixes
		if len(*buffer)-processed < 3 {
			// Not enough bytes for a complete prefix
			// Check if remaining bytes could be part of a prefix
			remainingBytes := (*buffer)[processed:]

			// If remaining bytes could be start of STDOUT_PREFIX (0x01, 0x01, 0x01)
			couldBeStdoutPrefix := true
			for i, b := range remainingBytes {
				if b != STDOUT_PREFIX[i] {
					couldBeStdoutPrefix = false
					break
				}
			}

			// If remaining bytes could be start of STDERR_PREFIX (0x02, 0x02, 0x02)
			couldBeStderrPrefix := true
			for i, b := range remainingBytes {
				if b != STDERR_PREFIX[i] {
					couldBeStderrPrefix = false
					break
				}
			}

			// If remaining bytes could be part of either prefix, keep them in buffer
			if couldBeStdoutPrefix || couldBeStderrPrefix {
				*buffer = remainingBytes
			} else {
				// Remaining bytes cannot be part of any prefix, output them
				result = append(result, remainingBytes...)
				*buffer = (*buffer)[:0]
			}
			break
		}

		// Check for STDOUT_PREFIX (0x01, 0x01, 0x01)
		if (*buffer)[processed] == STDOUT_PREFIX[0] &&
			(*buffer)[processed+1] == STDOUT_PREFIX[1] &&
			(*buffer)[processed+2] == STDOUT_PREFIX[2] {
			// Found STDOUT_PREFIX, skip it
			processed += 3
			continue
		}

		// Check for STDERR_PREFIX (0x02, 0x02, 0x02)
		if (*buffer)[processed] == STDERR_PREFIX[0] &&
			(*buffer)[processed+1] == STDERR_PREFIX[1] &&
			(*buffer)[processed+2] == STDERR_PREFIX[2] {
			// Found STDERR_PREFIX, skip it
			processed += 3
			continue
		}

		// No prefix found, add this byte to result
		result = append(result, (*buffer)[processed])
		processed++
	}

	// Remove processed bytes from buffer
	if processed > 0 && processed < len(*buffer) {
		*buffer = (*buffer)[processed:]
	}

	return result
}

// flushRemainingBuffer processes any remaining bytes in the buffer at the end of the stream
func flushRemainingBuffer(buffer *[]byte) []byte {
	if len(*buffer) == 0 {
		return nil
	}

	// At the end of stream, any remaining bytes are not prefixes (since they're incomplete)
	// So we should output them as regular data
	result := make([]byte, len(*buffer))
	copy(result, *buffer)
	*buffer = (*buffer)[:0] // Clear the buffer
	return result
}
