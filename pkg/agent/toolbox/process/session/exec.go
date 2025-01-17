// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SessionExecuteCommand(configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		sessionId := c.Param("sessionId")

		var request SessionExecuteRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.New("command is required"))
			return
		}

		session, ok := sessions[sessionId]
		if !ok {
			c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
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

		cmdToExec := fmt.Sprintf("\n%s ; echo \"DAYTONA_CMD_EXIT_CODE: $?\" ; echo DAYTONA_CMD_END\n", request.Command)

		output := make(chan string)
		outputErr := make(chan error)
		defer close(output)
		defer close(outputErr)

		go func() {
			out := ""
			for {
				line, _, err := session.OutReader.ReadLine()
				if err != nil {
					if err == io.EOF {
						out += string(line) + "EOF"
						break
					}
					outputErr <- err
				}
				if strings.Contains(string(line), "DAYTONA_CMD_END") {
					l := strings.Replace(string(line), "DAYTONA_CMD_END", "", 1)
					if request.Async {
						regex := regexp.MustCompile(`DAYTONA_CMD_EXIT_CODE: (\d+)`)
						l := regex.ReplaceAllString(l, "")
						logFile.Write([]byte(l))
					} else {
						out += l
					}
					break
				}

				l := string(line) + "\n"
				if request.Async {
					logFile.Write([]byte(l))
				} else {
					out += l
				}
			}

			if logFile != nil {
				logFile.Close()
			}
			output <- out
			outputErr <- nil
		}()

		_, err := session.StdinWriter.Write([]byte(cmdToExec))
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		var outputResult *string
		var exitCode *int
		select {
		case out := <-output:
			outputResult = &out

			exitCode, out = extractExitCode(*outputResult)
			outputResult = &out

			err := <-outputErr
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
		case err := <-outputErr:
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
		}

		c.JSON(http.StatusOK, SessionExecuteResponse{
			CommandId: cmdId,
			Output:    outputResult,
			ExitCode:  exitCode,
		})
	}
}

func extractExitCode(output string) (*int, string) {
	var exitCode *int

	regex := regexp.MustCompile(`DAYTONA_CMD_EXIT_CODE: (\d+)`)
	matches := regex.FindStringSubmatch(output)
	if len(matches) > 1 {
		code, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, output
		}
		exitCode = &code
	}

	if exitCode != nil {
		output = strings.Replace(output, fmt.Sprintf("DAYTONA_CMD_EXIT_CODE: %d", *exitCode), "", 1)
	}

	return exitCode, output
}
