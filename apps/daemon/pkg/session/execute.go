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
)

func (s *SessionService) Execute(sessionId, cmd string, async, isCombinedOutput, suppressInputEcho bool) (*SessionExecute, error) {
	session, ok := s.sessions.Get(sessionId)
	if !ok {
		return nil, common_errors.NewNotFoundError(errors.New("session not found"))
	}

	cmdId := uuid.NewString()

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

	defer logFile.Close()

	cmdToExec := fmt.Sprintf(cmdWrapperFormat+"\n",
		logFilePath, // %q  -> log
		logDir,      // %q  -> dir
		command.InputFilePath(session.Dir(s.configDir)), // %q  -> input
		toOctalEscapes(STDOUT_PREFIX),                   // %s  -> stdout prefix
		toOctalEscapes(STDERR_PREFIX),                   // %s  -> stderr prefix
		cmd,                                             // %s  -> verbatim script body
		exitCodeFilePath,                                // %q
	)

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

			exitCodeInt, err := strconv.Atoi(strings.TrimRight(string(exitCode), "\n"))
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

			logContent := string(logBytes)

			if isCombinedOutput {
				// remove prefixes from log bytes
				logBytes = bytes.ReplaceAll(bytes.ReplaceAll(logBytes, STDOUT_PREFIX, []byte{}), STDERR_PREFIX, []byte{})
				logContent = string(logBytes)
			}

			return &SessionExecute{
				CommandId: cmdId,
				Output:    &logContent,
				ExitCode:  &exitCodeInt,
			}, nil
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

var cmdWrapperFormat string = `
{
	log=%q
	dir=%q

	# per-command FIFOs
	sp="$dir/stdout.pipe"
	ep="$dir/stderr.pipe"
	ip=%q
	
	rm -f "$sp" "$ep" "$ip" && mkfifo "$sp" "$ep" "$ip" || exit 1

	cleanup() { rm -f "$sp" "$ep" "$ip"; }
	trap 'cleanup' EXIT HUP INT TERM

  # prefix each stream and append to shared log
	( while IFS= read -r line || [ -n "$line" ]; do printf '%s%%s\n' "$line"; done < "$sp" ) >> "$log" & r1=$!
	( while IFS= read -r line || [ -n "$line" ]; do printf '%s%%s\n' "$line"; done < "$ep" ) >> "$log" & r2=$!

	# Keep input FIFO open to prevent blocking when command opens stdin
	sleep infinity > "$ip" &

	# Run your command
	{ %s; } < "$ip" > "$sp" 2> "$ep"
	echo "$?" >> %s

	# drain labelers (cleanup via trap)
	wait "$r1" "$r2"

	# Ensure unlink even if the waits failed
	cleanup
}
`
