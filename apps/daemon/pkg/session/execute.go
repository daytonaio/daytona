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

func (s *SessionService) Execute(sessionId, cmd string, async, isCombinedOutput bool) (*SessionExecute, error) {
	session, ok := s.sessions[sessionId]
	if !ok {
		return nil, common_errors.NewNotFoundError(errors.New("session not found"))
	}

	// Strip whitespace and newlines from command to avoid issues
	cmd = strings.TrimSpace(cmd)

	cmdId := util.Pointer(uuid.NewString())

	command := &Command{
		Id:      *cmdId,
		Command: cmd,
	}
	session.commands[*cmdId] = command

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
			session.commands[*cmdId].ExitCode = util.Pointer(1)

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

			s.sessions[sessionId].commands[*cmdId].ExitCode = &exitCodeInt

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

	# Timeout-based pipe reader: reads line-by-line when available, flushes partial lines after 0.5s timeout
  read_pipe() {
    local pipe=$1
    local prefix=$2
    while true; do
      if IFS= read -r -t 0.5 line; then
        # Got a complete line with newline
        printf "$prefix%%s\n" "$line" >> "$log"
      else
        read_status=$?
        if [ $read_status -gt 128 ] && [ -n "$line" ]; then
          # Timeout occurred (status > 128) and we have partial data - flush it
          printf "$prefix%%s" "$line" >> "$log"
          line=""
        elif [ $read_status -eq 0 ] || [ $read_status -gt 128 ]; then
          # Either EOF (0) or timeout (>128) with no data - continue or break
          [ -p "$pipe" ] || break
        else
          # Pipe closed or error
          break
        fi
      fi
    done < "$pipe"
  }

  # Start readers for stdout and stderr
  read_pipe "$sp" '%s' & r1=$!
  read_pipe "$ep" '%s' & r2=$!

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
