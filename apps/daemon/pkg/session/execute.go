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

func (s *SessionService) Execute(sessionId, cmdId, cmd string, async, isCombinedOutput, suppressInputEcho bool) (*SessionExecute, error) {
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

	defer logFile.Close()

	inputPipeCommand := `cat /dev/null > "$ip" &`
	if async {
		inputPipeCommand = `while :; do sleep 3600; done > "$ip" &`
	}

	cmdToExec := fmt.Sprintf(cmdWrapperFormat+"\n",
		logFilePath, // %q  -> log
		logDir,      // %q  -> dir
		command.InputFilePath(session.Dir(s.configDir)), // %q  -> input
		toOctalEscapes(log.STDOUT_PREFIX),               // %s  -> stdout prefix
		toOctalEscapes(log.STDERR_PREFIX),               // %s  -> stderr prefix
		inputPipeCommand,                                // %s  -> stdin behavior
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

			if isCombinedOutput {
				stripped := bytes.ReplaceAll(bytes.ReplaceAll(logBytes, log.STDOUT_PREFIX, []byte{}), log.STDERR_PREFIX, []byte{})
				output := string(stripped)
				return &SessionExecute{
					CommandId: cmdId,
					Output:    &output,
					ExitCode:  &exitCodeInt,
				}, nil
			}

			stdoutBytes, stderrBytes := demuxLogBytes(logBytes)
			output := string(logBytes)
			stdoutStr := string(stdoutBytes)
			stderrStr := string(stderrBytes)
			return &SessionExecute{
				CommandId: cmdId,
				Output:    &output,
				Stdout:    &stdoutStr,
				Stderr:    &stderrStr,
				ExitCode:  &exitCodeInt,
			}, nil
		}
	}
}

func demuxLogBytes(data []byte) (stdout, stderr []byte) {
	var outBuf, errBuf []byte
	pos := 0

	for pos < len(data) {
		outIdx := bytes.Index(data[pos:], log.STDOUT_PREFIX)
		errIdx := bytes.Index(data[pos:], log.STDERR_PREFIX)

		if outIdx >= 0 {
			outIdx += pos
		}
		if errIdx >= 0 {
			errIdx += pos
		}

		nextIdx := -1
		var nextIsStdout bool
		if outIdx >= 0 && (errIdx < 0 || outIdx <= errIdx) {
			nextIdx = outIdx
			nextIsStdout = true
		} else if errIdx >= 0 {
			nextIdx = errIdx
			nextIsStdout = false
		}

		if nextIdx < 0 {
			break
		}

		endOfPrefix := nextIdx + len(log.STDOUT_PREFIX)

		nextMarker := len(data)
		for search := endOfPrefix; search < len(data); {
			oi := bytes.Index(data[search:], log.STDOUT_PREFIX)
			ei := bytes.Index(data[search:], log.STDERR_PREFIX)
			if oi >= 0 {
				oi += search
			}
			if ei >= 0 {
				ei += search
			}

			candidate := -1
			if oi >= 0 && (ei < 0 || oi <= ei) {
				candidate = oi
			} else if ei >= 0 {
				candidate = ei
			}

			if candidate < 0 {
				break
			}
			nextMarker = candidate
			break
		}

		chunk := data[endOfPrefix:nextMarker]
		if nextIsStdout {
			outBuf = append(outBuf, chunk...)
		} else {
			errBuf = append(errBuf, chunk...)
		}

		pos = nextMarker
	}

	return outBuf, errBuf
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

	# Sync commands should see EOF immediately; async commands keep stdin open for SendInput.
	%s
	ip_pid=$!

	# Run your command
	{ %s; } < "$ip" > "$sp" 2> "$ep"
	echo "$?" >> %s

	# Stop the stdin holder so it doesn't outlive the command
	kill "$ip_pid" 2>/dev/null; wait "$ip_pid" 2>/dev/null

	# drain labelers (cleanup via trap)
	wait "$r1" "$r2"

	# Ensure unlink even if the waits failed
	cleanup
}
`
