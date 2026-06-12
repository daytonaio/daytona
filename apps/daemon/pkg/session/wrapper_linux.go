//go:build !windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/daytonaio/common-go/pkg/log"
)

// buildCommandInvocation persists the per-command artifacts under logDir and
// returns the line(s) written to the session shell's stdin to run them.
func buildCommandInvocation(logDir, logFilePath, exitCodeFilePath, inputFilePath, cmd string, async bool) (string, error) {
	cmdFilePath := filepath.Join(logDir, "cmd.sh")
	if err := os.WriteFile(cmdFilePath, []byte(cmd), 0600); err != nil {
		return "", fmt.Errorf("failed to write command file: %w", err)
	}

	inputPipeCommand := `cat /dev/null > "$ip" &`
	if async {
		inputPipeCommand = `while :; do sleep 3600; done > "$ip" &`
	}

	return fmt.Sprintf(cmdWrapperFormat+"\n",
		logFilePath,                       // %q  -> log
		logDir,                            // %q  -> dir
		inputFilePath,                     // %q  -> input
		toOctalEscapes(log.STDOUT_PREFIX), // %s  -> stdout prefix
		toOctalEscapes(log.STDERR_PREFIX), // %s  -> stderr prefix
		inputPipeCommand,                  // %s  -> stdin behavior
		cmdFilePath,                       // %q  -> command file path
		exitCodeFilePath,                  // %q
	), nil
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

	# Run your command from file (avoids heredoc parsing issues with pipe-fed shells)
	{ . %q; } < "$ip" > "$sp" 2> "$ep"
	_ec=$?

	# Stop the stdin holder so it doesn't outlive the command
	kill "$ip_pid" 2>/dev/null; wait "$ip_pid" 2>/dev/null

	# drain labelers (cleanup via trap)
	wait "$r1" "$r2"

	# Write exit code only after labelers have flushed all output to the log file.
	# Previously echo "$?" ran before wait, creating a race where clients polling
	# the exit-code file would read an empty/incomplete log.
	echo "$_ec" >> %q

	# Ensure unlink even if the waits failed
	cleanup
}
`
