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
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	log "github.com/sirupsen/logrus"
)

func (s *SessionController) GetSessionCommandLogs(c *gin.Context) {
	sessionId := c.Param("sessionId")
	cmdId := c.Param("commandId")

	session, ok := sessions[sessionId]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
		return
	}

	command, ok := sessions[sessionId].commands[cmdId]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("command not found"))
		return
	}

	logFilePath, exitCodeFilePath := command.LogFilePath(session.Dir(s.configDir))

	sdkVersion := util.ExtractSdkVersionFromHeader(c.Request.Header)
	if sdkVersion != "" {
		upgrader.Subprotocols = []string{"X-Daytona-SDK-Version~" + sdkVersion}
	} else {
		upgrader.Subprotocols = []string{}
	}

	versionComparison, err := util.CompareVersions(sdkVersion, "0.27.0-0")
	if err != nil {
		log.Debug(err)
		versionComparison = util.Pointer(1)
	}
	isLegacy := versionComparison != nil && *versionComparison < 0 && sdkVersion != "0.0.0-dev"

	if c.Request.Header.Get("Upgrade") == "websocket" {
		logFile, err := os.Open(logFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				c.AbortWithError(http.StatusNotFound, err)
				return
			}
			if os.IsPermission(err) {
				c.AbortWithError(http.StatusForbidden, err)
				return
			}
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		defer logFile.Close()
		ReadLog(c, logFile, util.ReadLogWithExitCode, exitCodeFilePath, func(conn *websocket.Conn, messages chan []byte, errors chan error) {
			var buffer []byte
			for {
				select {
				case <-session.ctx.Done():
					// Flush any remaining bytes in buffer before closing
					if isLegacy && len(buffer) > 0 {
						remainingData := flushRemainingBuffer(&buffer)
						if len(remainingData) > 0 {
							err := conn.WriteMessage(websocket.TextMessage, remainingData)
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
					if isLegacy {
						// Process chunks with buffering to handle prefixes split across chunks
						processedData := processLogChunkWithPrefixFiltering(msg, &buffer)
						if len(processedData) > 0 {
							err := conn.WriteMessage(websocket.TextMessage, processedData)
							if err != nil {
								errors <- err
								return
							}
						}
					} else {
						err := conn.WriteMessage(websocket.TextMessage, msg)
						if err != nil {
							errors <- err
							return
						}
					}
				case <-errors:
					// Stream ended, flush any remaining bytes in buffer
					if isLegacy && len(buffer) > 0 {
						remainingData := flushRemainingBuffer(&buffer)
						if len(remainingData) > 0 {
							writeErr := conn.WriteMessage(websocket.TextMessage, remainingData)
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
		return
	}

	logBytes, err := os.ReadFile(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		if os.IsPermission(err) {
			c.AbortWithError(http.StatusForbidden, err)
			return
		}
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if isLegacy {
		// remove prefixes from log bytes for backwards compatibility
		logBytes = bytes.ReplaceAll(bytes.ReplaceAll(logBytes, STDOUT_PREFIX, []byte{}), STDERR_PREFIX, []byte{})
	}

	c.String(http.StatusOK, string(logBytes))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ReadLog reads from the logReader and writes to the websocket.
// T is the type of the message to be read from the logReader
func ReadLog[T any](ginCtx *gin.Context, logReader io.Reader, readFunc func(context.Context, io.Reader, bool, string, chan T, chan error), exitCodeFilePath string, wsWriteFunc func(*websocket.Conn, chan T, chan error)) {
	followQuery := ginCtx.Query("follow")
	follow := followQuery == "true"

	ws, err := upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
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

	msgChannel := make(chan T)
	errChannel := make(chan error)
	ctx, cancel := context.WithCancel(ginCtx.Request.Context())

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
	i := 0

	for i < len(*buffer) {
		// Check if we have enough bytes to check for prefixes
		if len(*buffer)-i < 3 {
			// Not enough bytes for a complete prefix, keep remaining bytes in buffer
			*buffer = (*buffer)[i:]
			break
		}

		// Check for STDOUT_PREFIX (0x01, 0x01, 0x01)
		if (*buffer)[i] == STDOUT_PREFIX[0] && (*buffer)[i+1] == STDOUT_PREFIX[1] && (*buffer)[i+2] == STDOUT_PREFIX[2] {
			// Found STDOUT_PREFIX, skip it
			i += 3
			continue
		}

		// Check for STDERR_PREFIX (0x02, 0x02, 0x02)
		if (*buffer)[i] == STDERR_PREFIX[0] && (*buffer)[i+1] == STDERR_PREFIX[1] && (*buffer)[i+2] == STDERR_PREFIX[2] {
			// Found STDERR_PREFIX, skip it
			i += 3
			continue
		}

		// No prefix found, add this byte to result
		result = append(result, (*buffer)[i])
		i++
	}

	// If we processed all bytes, clear the buffer
	if i >= len(*buffer) {
		*buffer = (*buffer)[:0]
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
