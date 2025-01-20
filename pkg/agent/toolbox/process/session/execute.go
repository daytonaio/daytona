// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"fmt"
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
			c.AbortWithError(http.StatusBadRequest, err)
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

			command := &Command{
				Id:      *cmdId,
				Command: request.Command,
			}
			session.commands[*cmdId] = command

			err := os.MkdirAll(filepath.Join(configDir, "sessions", sessionId, *cmdId), 0755)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			logFile, err = os.Create(filepath.Join(configDir, "sessions", sessionId, *cmdId, "output.log"))
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
		}

		cmdToExec := fmt.Sprintf("\n%s ; echo \"DAYTONA_CMD_EXIT_CODE: $?\"\n", request.Command)

		type execResult struct {
			out      string
			err      error
			exitCode *int
		}

		resultChan := make(chan execResult)

		go func() {
			out := ""
			var exitCode *int
			defer close(resultChan)

			for session.outReader.Scan() {
				line := session.outReader.Text()
				line = line + "\n"

				exitCode, line = extractExitCode(line)

				if request.Async {
					_, err := logFile.Write([]byte(line))
					if err != nil {
						resultChan <- execResult{err: err}
						return
					}
				} else {
					out += line
				}

				if exitCode != nil {
					if request.Async {
						sessions[sessionId].commands[*cmdId].ExitCode = exitCode
					}
					break
				}
			}

			if logFile != nil {
				logFile.Close()
			}
			resultChan <- execResult{out: out, exitCode: exitCode, err: session.outReader.Err()}
		}()

		_, err := session.stdinWriter.Write([]byte(cmdToExec))
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if request.Async {
			c.JSON(http.StatusAccepted, SessionExecuteResponse{
				CommandId: cmdId,
			})
			return
		}

		result := <-resultChan
		if result.err != nil {
			c.AbortWithError(http.StatusBadRequest, result.err)
			return
		}

		c.JSON(http.StatusOK, SessionExecuteResponse{
			CommandId: cmdId,
			Output:    &result.out,
			ExitCode:  result.exitCode,
		})
	}
}

func extractExitCode(output string) (*int, string) {
	var exitCode *int

	regex := regexp.MustCompile(`DAYTONA_CMD_EXIT_CODE: (\d+)\n`)
	matches := regex.FindStringSubmatch(output)
	if len(matches) > 1 {
		code, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, output
		}
		exitCode = &code
	}

	if exitCode != nil {
		output = strings.Replace(output, fmt.Sprintf("DAYTONA_CMD_EXIT_CODE: %d\n", *exitCode), "", 1)
	}

	return exitCode, output
}
