// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/google/uuid"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/log"
)

func (s *SessionService) Execute(sessionId, cmdId, cmd string, async, isCombinedOutput, skipServerDemux, suppressInputEcho bool) (*SessionExecute, error) {
	session, ok := s.sessions.Get(sessionId)
	if !ok {
		return nil, common_errors.NewNotFoundError(errors.New("session not found"))
	}

	if cmdId == util.EmptyCommandID {
		cmdId = uuid.NewString()
	}

	if _, ok := session.commands.Get(cmdId); ok {
		return nil, common_errors.NewConflictError(errors.New("command with the given ID already exists"))
	}

	command := &Command{
		Id:                cmdId,
		Command:           cmd,
		SuppressInputEcho: suppressInputEcho,
	}
	session.commands.Set(cmdId, command)

	logFilePath, exitCodeFilePath := command.LogFilePath(session.Dir(s.configDir))
	logDir := filepath.Dir(logFilePath)

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, common_errors.NewBadRequestError(fmt.Errorf("failed to create log directory: %w", err))
	}

	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, common_errors.NewBadRequestError(fmt.Errorf("failed to create log file: %w", err))
	}
	// Close immediately: the handle exists only to create/truncate the file.
	// Holding it across the poll loop blocks Windows writers (sharing violation)
	// and Delete's RemoveAll.
	logFile.Close()

	cmdToExec, err := buildCommandInvocation(logDir, logFilePath, exitCodeFilePath, command.InputFilePath(session.Dir(s.configDir)), cmd, async)
	if err != nil {
		return nil, common_errors.NewBadRequestError(err)
	}

	_, err = session.stdinWriter.Write([]byte(cmdToExec))
	if err != nil {
		return nil, common_errors.NewBadRequestError(fmt.Errorf("failed to write command: %w", err))
	}

	if async {
		return &SessionExecute{
			CommandId: cmdId,
		}, nil
	}

	for {
		select {
		case <-session.ctx.Done():
			command, ok := session.commands.Get(cmdId)
			if !ok {
				return nil, common_errors.NewBadRequestError(errors.New("command not found"))
			}

			command.ExitCode = util.Pointer(1)

			return nil, common_errors.NewBadRequestError(errors.New("session cancelled"))
		default:
			exitCode, err := os.ReadFile(exitCodeFilePath)
			if err != nil {
				if os.IsNotExist(err) {
					time.Sleep(50 * time.Millisecond)
					continue
				}
				return nil, common_errors.NewBadRequestError(fmt.Errorf("failed to read exit code file: %w", err))
			}

			// The writer creates the exit-code file before flushing its content;
			// an empty read means the write is still in flight, not a result.
			if strings.TrimSpace(string(exitCode)) == "" {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			exitCodeInt, err := strconv.Atoi(strings.TrimSpace(string(exitCode)))
			if err != nil {
				return nil, common_errors.NewBadRequestError(fmt.Errorf("failed to convert exit code to int: %w", err))
			}

			command, ok := session.commands.Get(cmdId)
			if !ok {
				return nil, common_errors.NewBadRequestError(errors.New("command not found"))
			}
			command.ExitCode = &exitCodeInt

			logBytes, err := os.ReadFile(logFilePath)
			if err != nil {
				return nil, common_errors.NewBadRequestError(fmt.Errorf("failed to read log file: %w", err))
			}

			if isCombinedOutput {
				stripped := bytes.ReplaceAll(bytes.ReplaceAll(logBytes, log.STDOUT_PREFIX, []byte{}), log.STDERR_PREFIX, []byte{})
				output := string(stripped)
				return &SessionExecute{
					CommandId: cmdId,
					Output:    &output,
					ExitCode:  &exitCodeInt,
				}, nil
			}

			output := string(logBytes)
			result := &SessionExecute{
				CommandId: cmdId,
				Output:    &output,
				ExitCode:  &exitCodeInt,
			}

			if !skipServerDemux {
				stdoutBytes, stderrBytes := DemuxLogBytes(logBytes)
				stdoutStr := string(stdoutBytes)
				stderrStr := string(stderrBytes)
				result.Stdout = &stdoutStr
				result.Stderr = &stderrStr
			}

			return result, nil
		}
	}
}

func DemuxLogBytes(data []byte) (stdout, stderr []byte) {
	prefixLen := len(log.STDOUT_PREFIX)
	var outBuf, errBuf []byte
	pos := 0

	for pos < len(data) {
		if pos+prefixLen <= len(data) && bytes.Equal(data[pos:pos+prefixLen], log.STDOUT_PREFIX) {
			end := findNextMarker(data, pos+prefixLen, prefixLen)
			outBuf = append(outBuf, data[pos+prefixLen:end]...)
			pos = end
		} else if pos+prefixLen <= len(data) && bytes.Equal(data[pos:pos+prefixLen], log.STDERR_PREFIX) {
			end := findNextMarker(data, pos+prefixLen, prefixLen)
			errBuf = append(errBuf, data[pos+prefixLen:end]...)
			pos = end
		} else {
			pos++
		}
	}

	return outBuf, errBuf
}

func findNextMarker(data []byte, from int, prefixLen int) int {
	for i := from; i+prefixLen <= len(data); i++ {
		if bytes.Equal(data[i:i+prefixLen], log.STDOUT_PREFIX) || bytes.Equal(data[i:i+prefixLen], log.STDERR_PREFIX) {
			return i
		}
	}
	return len(data)
}
