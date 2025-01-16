// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Session struct {
	Cmd          *exec.Cmd
	StdoutReader *bufio.Reader
	StdinWriter  io.Writer
}

var sessions = map[string]*Session{}

func CreateSession(configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		cmd := exec.Command("/bin/sh")

		var request CreateSessionRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.AbortWithError(400, fmt.Errorf("invalid request body: %w", err))
			return
		}

		stdinWriter, err := cmd.StdinPipe()
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		stdoutReader, err := cmd.StdoutPipe()
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		err = cmd.Start()
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		err = os.MkdirAll(filepath.Join(configDir, "sessions", request.SessionId), 0755)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		sessions[request.SessionId] = &Session{
			Cmd:          cmd,
			StdoutReader: bufio.NewReader(stdoutReader),
			StdinWriter:  stdinWriter,
		}

		c.Status(201)
	}
}

func SessionExecuteCommand(configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		sessionId := c.Param("sessionId")

		var request SessionExecuteRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.AbortWithError(400, errors.New("command is required"))
			return
		}

		var cmdId *string
		var logFile *os.File

		if request.Async {
			cmdId = util.Pointer(uuid.NewString())
			err := os.MkdirAll(filepath.Join(configDir, "sessions", sessionId, *cmdId), 0755)
			if err != nil {
				c.AbortWithError(400, err)
				return
			}

			logFile, err = os.Create(filepath.Join(configDir, "sessions", sessionId, *cmdId, "output.log"))
			if err != nil {
				c.AbortWithError(400, err)
				return
			}
		}

		session, ok := sessions[sessionId]
		if !ok {
			c.AbortWithError(404, errors.New("session not found"))
			return
		}

		cmdToExec := fmt.Sprintf("%s && echo DAYTONA_CMD_END\n\n", request.Command)

		output := make(chan string)
		outputErr := make(chan error)
		defer close(output)
		defer close(outputErr)

		go func() {
			out := ""
			for {
				line, _, err := session.StdoutReader.ReadLine()
				if err != nil {
					if err == io.EOF {
						break
					}
					outputErr <- err
				}
				if strings.Contains(string(line), "DAYTONA_CMD_END") {
					break
				}

				l := string(line) + "\n"
				if request.Async {
					logFile.Write([]byte(l))
				} else {
					out += l
				}
			}

			logFile.Close()
			output <- out
			outputErr <- nil
		}()

		_, err := session.StdinWriter.Write([]byte(cmdToExec))
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		var outputResult *string
		select {
		case out := <-output:
			outputResult = &out
			err := <-outputErr
			if err != nil {
				c.AbortWithError(400, err)
				return
			}
		case err := <-outputErr:
			if err != nil {
				c.AbortWithError(400, err)
				return
			}
		}

		c.JSON(200, SessionExecuteResponse{
			CommandId: cmdId,
			Output:    outputResult,
		})
	}
}

func DeleteSession(configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		sessionId := c.Param("sessionId")

		session, ok := sessions[sessionId]
		if !ok {
			c.AbortWithError(404, errors.New("session not found"))
			return
		}

		err := session.Cmd.Process.Kill()
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		delete(sessions, sessionId)

		err = os.RemoveAll(filepath.Join(configDir, "sessions", sessionId))
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		c.Status(204)
	}
}
