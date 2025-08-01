// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
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
	if !ok || session.deleted {
		c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
		return
	}

	command, ok := sessions[sessionId].commands[cmdId]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("command not found"))
		return
	}

	stdoutPath, stderrPath, _ := command.LogFilePath(session.Dir(s.configDir))

	fmt.Println("[DEBUG] header", c.Request.Header)

	sdkVersion := c.Request.Header.Get("X-Daytona-SDK-Version")
	versionComparison, err := util.CompareVersions(sdkVersion, "0.26.0-0")
	if err != nil {
		log.Error(err)
		versionComparison = 1
	}
	isLegacy := versionComparison < 0 && sdkVersion != "0.0.0-dev" && strings.Contains(c.Request.Header.Get("X-Daytona-Source"), "sdk")

	if c.Request.Header.Get("Upgrade") == "websocket" {
		stdoutFile, err := os.Open(stdoutPath)
		if err != nil {
			handleLogFileError(c, err)
			return
		}
		defer stdoutFile.Close()

		stderrFile, err := os.Open(stderrPath)
		if err != nil {
			handleLogFileError(c, err)
			return
		}
		defer stderrFile.Close()

		cleanupConn := func(conn *websocket.Conn) {
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				time.Now().Add(time.Second),
			)
			conn.Close()
		}

		StdMux(c, isLegacy, stdoutFile, stderrFile, util.ReadCommandLog, func(conn *websocket.Conn, messages chan []byte, errors chan error) {
			for {
				select {
				case <-session.ctx.Done():
					cleanupConn(conn)
					return
				case msg, ok := <-messages:
					if !ok { // channel is closed
						cleanupConn(conn)
						return
					}
					err := conn.WriteMessage(websocket.BinaryMessage, msg)
					if err != nil {
						errors <- err
						return
					}
				}
			}
		})
		return
	}

	stdoutBytes, err := os.ReadFile(stdoutPath)
	if err != nil {
		handleLogFileError(c, err)
		return
	}
	stdoutContent := strings.TrimSuffix(strings.TrimRight(string(stdoutBytes), " \n\r\t"), COMMAND_EXIT_MARKER)

	stderrBytes, err := os.ReadFile(stderrPath)
	if err != nil {
		handleLogFileError(c, err)
		return
	}
	stderrContent := strings.TrimSuffix(strings.TrimRight(string(stderrBytes), " \n\r\t"), COMMAND_EXIT_MARKER)

	if isLegacy {
		fmt.Println("[DEBUG] answering with legacy format")
		c.JSON(http.StatusOK, stdoutContent+"\n"+stderrContent)
	} else {
		fmt.Println("[DEBUG] answering with new format")
		c.JSON(http.StatusOK, SessionCommandLogsResponse{
			Stdout: stdoutContent,
			Stderr: stderrContent,
		})
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// StdMux reads from the logReader and writes to the websocket.
// T is the type of the message to be read from the logReader
func StdMux(ginCtx *gin.Context, isLegacy bool, stdoutReader io.Reader, stderrReader io.Reader, readFunc func(context.Context, io.Reader, bool, chan []byte, chan error), wsWriteFunc func(*websocket.Conn, chan []byte, chan error)) {
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

	msgChannel := make(chan []byte)
	stdoutChannel := make(chan []byte)
	stderrChannel := make(chan []byte)
	errChannel := make(chan error)
	ctx, cancel := context.WithCancel(ginCtx.Request.Context())

	defer cancel()
	go readFunc(ctx, stdoutReader, follow, stdoutChannel, errChannel)
	go readFunc(ctx, stderrReader, follow, stderrChannel, errChannel)
	go wsWriteFunc(ws, msgChannel, errChannel)

	streams := []struct {
		channel chan []byte
		prefix  byte
	}{
		{stdoutChannel, 0x01},
		{stderrChannel, 0x02},
	}

	go func() {
		var wg sync.WaitGroup
		wg.Add(2)

		if isLegacy {
			for _, stream := range streams {
				go func(ch chan []byte) {
					defer wg.Done()
					for data := range ch {
						if idx := bytes.Index(data, []byte(COMMAND_EXIT_MARKER)); idx != -1 {
							// Send everything up to (but not including) the marker
							data = data[:idx]
							if len(data) > 0 {
								msgChannel <- data
							}
							close(ch)
							return
						}
						if len(data) > 0 {
							msgChannel <- data
						}
					}
				}(stream.channel)
			}
		} else {
			for _, stream := range streams {
				go func(ch chan []byte, prefix byte) {
					defer wg.Done()
					for data := range ch {
						if idx := bytes.Index(data, []byte(COMMAND_EXIT_MARKER)); idx != -1 {
							// Send everything up to (but not including) the marker
							data = data[:idx]
							if len(data) > 0 {
								msgChannel <- append([]byte{prefix}, data...)
							}
							close(ch)
							return
						}
						if len(data) > 0 {
							msgChannel <- append([]byte{prefix}, data...)
						}
					}
				}(stream.channel, stream.prefix)
			}
		}

		wg.Wait()
		close(msgChannel)
	}()

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

func handleLogFileError(c *gin.Context, err error) {
	if os.IsNotExist(err) {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	if os.IsPermission(err) {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}
	c.AbortWithError(http.StatusBadRequest, err)
}
