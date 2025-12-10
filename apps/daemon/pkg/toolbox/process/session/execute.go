// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

var (
	STDOUT_PREFIX = []byte{0x01, 0x01, 0x01}
	STDERR_PREFIX = []byte{0x02, 0x02, 0x02}
)

// Add a standard error response struct
type ErrorResponse struct {
	Error string `json:"error"`
}

// SessionExecuteCommand godoc
//
//	@Summary		Execute command in session
//	@Description	Execute a command within an existing shell session
//	@Tags			process
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string					true	"Session ID"
//	@Param			request		body		SessionExecuteRequest	true	"Command execution request"
//	@Success		200			{object}	SessionExecuteResponse
//	@Success		202			{object}	SessionExecuteResponse
//	@Router			/process/session/{sessionId}/exec [post]
//
//	@id				SessionExecuteCommand
func (s *SessionController) SessionExecuteCommand(c *gin.Context) {
	sessionId := c.Param("sessionId")

	var request SessionExecuteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if request.Async {
		request.RunAsync = true
	}

	// Validate command is not empty (if not already handled by binding)
	if strings.TrimSpace(request.Command) == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("command cannot be empty"))
		return
	}

	session, ok := sessions[sessionId]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
		return
	}

	cmdId := uuid.NewString()

	command := &Command{
		Id:      cmdId,
		Command: request.Command,
	}
	session.commands[cmdId] = command

	logFilePath, exitCodeFilePath := command.LogFilePath(session.Dir(s.configDir))
	logDir := filepath.Dir(logFilePath)

	if err := os.MkdirAll(logDir, 0755); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to create log directory: %w", err))
		return
	}

	logFile, err := os.Create(logFilePath)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to create log file: %w", err))
		return
	}

	defer logFile.Close()

	cmdToExec := fmt.Sprintf(
		`{
	log=%q
	dir=%q

	# per-command FIFOs
	sp="$dir/stdout.pipe.%s.$$"; ep="$dir/stderr.pipe.%s.$$"
	rm -f "$sp" "$ep" && mkfifo "$sp" "$ep" || exit 1

	cleanup() { rm -f "$sp" "$ep"; }
	trap 'cleanup' EXIT HUP INT TERM

	# prefix each stream and append to shared log
	( while IFS= read -r line || [ -n "$line" ]; do printf '%s%%s\n' "$line"; done < "$sp" ) >> "$log" & r1=$!
	( while IFS= read -r line || [ -n "$line" ]; do printf '%s%%s\n' "$line"; done < "$ep" ) >> "$log" & r2=$!

	# Run your command
	{ %s; } > "$sp" 2> "$ep"
	echo "$?" >> %s

	# drain labelers (cleanup via trap)
	wait "$r1" "$r2"

	# Ensure unlink even if the waits failed
	cleanup
}
`+"\n",
		logFilePath,  // %q  -> log
		logDir,       // %q  -> dir
		cmdId, cmdId, // %s  %s -> fifo names
		toOctalEscapes(STDOUT_PREFIX), // %s  -> stdout prefix
		toOctalEscapes(STDERR_PREFIX), // %s  -> stderr prefix
		request.Command,               // %s  -> verbatim script body
		exitCodeFilePath,              // %q
	)

	_, err = session.stdinWriter.Write([]byte(cmdToExec))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to write command: %w", err))
		return
	}

	if request.RunAsync {
		c.JSON(http.StatusAccepted, SessionExecuteResponse{
			CommandId: cmdId,
		})
		return
	}

	for {
		select {
		case <-session.ctx.Done():
			session.commands[cmdId].ExitCode = util.Pointer(1)

			c.AbortWithError(http.StatusBadRequest, errors.New("session cancelled"))
			return
		default:
			exitCode, err := os.ReadFile(exitCodeFilePath)
			if err != nil {
				if os.IsNotExist(err) {
					time.Sleep(50 * time.Millisecond)
					continue
				}
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to read exit code file: %w", err))
				return
			}

			exitCodeInt, err := strconv.Atoi(strings.TrimRight(string(exitCode), "\n"))
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to convert exit code to int: %w", err))
				return
			}

			sessions[sessionId].commands[cmdId].ExitCode = &exitCodeInt

			logBytes, err := os.ReadFile(logFilePath)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to read log file: %w", err))
				return
			}
			logContent := string(logBytes)

			sdkVersion := util.ExtractSdkVersionFromHeader(c.Request.Header)
			if sdkVersion != "" {
				upgrader.Subprotocols = []string{"X-Daytona-SDK-Version~" + sdkVersion}
			} else {
				upgrader.Subprotocols = []string{}
			}

			versionComparison, err := util.CompareVersions(sdkVersion, "0.27.0-0")
			if err != nil {
				log.Error(err)
				versionComparison = util.Pointer(1)
			}
			isCombinedOutput := (versionComparison != nil && *versionComparison < 0 && sdkVersion != "0.0.0-dev") || (sdkVersion == "" && c.Request.Header.Get("X-Daytona-Split-Output") != "true")

			if isCombinedOutput {
				// remove prefixes from log bytes
				logBytes = bytes.ReplaceAll(bytes.ReplaceAll(logBytes, STDOUT_PREFIX, []byte{}), STDERR_PREFIX, []byte{})
				logContent = string(logBytes)
			}

			c.JSON(http.StatusOK, SessionExecuteResponse{
				CommandId: cmdId,
				Output:    &logContent,
				ExitCode:  &exitCodeInt,
			})
			return
		}
	}
}

func toOctalEscapes(b []byte) string {
	out := ""
	for _, c := range b {
		out += fmt.Sprintf("\\%03o", c) // e.g. 0x01 → \001
	}
	return out
}
