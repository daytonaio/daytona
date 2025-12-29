// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/daytonaio/daemon-win/internal/util"
	"github.com/daytonaio/daemon-win/pkg/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

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

	// Validate command is not empty
	if strings.TrimSpace(request.Command) == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("command cannot be empty"))
		return
	}

	// Parse Linux shell wrapper (sh -c "...") and extract actual command
	parsedCommand, envVars := common.ParseShellWrapper(request.Command)
	if parsedCommand != request.Command {
		log.Debugf("Parsed shell wrapper: %q -> %q (env: %v)", request.Command, parsedCommand, envVars)
	}

	// Build Windows command with env vars if any
	finalCommand := common.BuildWindowsCommand(parsedCommand, envVars)

	session, ok := sessions[sessionId]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
		return
	}

	cmdId := util.Pointer(uuid.NewString())

	command := &Command{
		Id:      *cmdId,
		Command: request.Command,
	}
	session.commands[*cmdId] = command

	logFilePath, exitCodeFilePath := command.LogFilePath(session.Dir(s.configDir))
	logDir := filepath.Dir(logFilePath)

	if err := os.MkdirAll(logDir, 0755); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to create log directory: %w", err))
		return
	}

	// Windows PowerShell command execution script
	// This captures both stdout and stderr to the log file and writes exit code
	// We use [Console]::Out to avoid BOM issues with Out-File
	cmdToExec := fmt.Sprintf(`
$logFile = "%s"
$exitCodeFile = "%s"
$exitCode = 0
try {
    $output = & { %s } 2>&1
    if ($output) {
        $output | Out-File -FilePath $logFile -Encoding UTF8 -Append
    }
    if ($LASTEXITCODE -ne $null) {
        $exitCode = $LASTEXITCODE
    }
} catch {
    $_.Exception.Message | Out-File -FilePath $logFile -Encoding UTF8 -Append
    $exitCode = 1
}
[System.IO.File]::WriteAllText($exitCodeFile, $exitCode.ToString())
`,
		strings.ReplaceAll(logFilePath, `\`, `\\`),
		strings.ReplaceAll(exitCodeFilePath, `\`, `\\`),
		finalCommand,
	)

	_, err := session.stdinWriter.Write([]byte(cmdToExec + "\n"))
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

	// Wait for command completion by polling for exit code file
	for {
		select {
		case <-session.ctx.Done():
			session.commands[*cmdId].ExitCode = util.Pointer(1)
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

			// Clean up the exit code string
			exitCodeStr := strings.TrimSpace(string(exitCode))
			exitCodeStr = strings.TrimRight(exitCodeStr, "\r\n")
			// Remove UTF-8 BOM if present
			exitCodeStr = strings.TrimPrefix(exitCodeStr, "\xef\xbb\xbf")
			exitCodeStr = strings.TrimSpace(exitCodeStr)

			// If empty, command might still be running, wait more
			if exitCodeStr == "" {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			exitCodeInt, err := strconv.Atoi(exitCodeStr)
			if err != nil {
				log.Warnf("Failed to parse exit code '%s': %v", exitCodeStr, err)
				exitCodeInt = 1 // Default to error
			}

			sessions[sessionId].commands[*cmdId].ExitCode = &exitCodeInt

			logBytes, err := os.ReadFile(logFilePath)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to read log file: %w", err))
				return
			}
			logContent := string(logBytes)

			sdkVersion := util.ExtractSdkVersionFromHeader(c.Request.Header)
			versionComparison, err := util.CompareVersions(sdkVersion, "0.27.0-0")
			if err != nil {
				log.Error(err)
				versionComparison = util.Pointer(1)
			}
			_ = versionComparison // For now, we use combined output on Windows

			c.JSON(http.StatusOK, SessionExecuteResponse{
				CommandId: cmdId,
				Output:    &logContent,
				ExitCode:  &exitCodeInt,
			})
			return
		}
	}
}
