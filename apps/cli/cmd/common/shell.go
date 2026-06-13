// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import "strings"

// ShellJoinArgs joins command arguments into a single shell command line.
// A single argument is returned verbatim so already-quoted commands
// (e.g. `daytona exec my-sandbox "echo hi | wc -l"`) keep working. With
// multiple arguments each one is quoted when needed so the remote shell
// sees the same argument vector the user typed locally.
func ShellJoinArgs(args []string) string {
	if len(args) == 1 {
		return args[0]
	}

	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuoteArg(arg)
	}
	return strings.Join(quoted, " ")
}

func shellQuoteArg(arg string) string {
	if arg == "" {
		return "''"
	}
	if !shellArgNeedsQuoting(arg) {
		return arg
	}
	return "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
}

// shellArgNeedsQuoting reports whether arg contains any byte outside the
// POSIX-safe set [A-Za-z0-9_./:=@%^,+-].
func shellArgNeedsQuoting(arg string) bool {
	for i := 0; i < len(arg); i++ {
		c := arg[i]
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '_' || c == '.' || c == '/' || c == ':' || c == '=' ||
			c == '@' || c == '%' || c == '^' || c == ',' || c == '+' || c == '-':
		default:
			return true
		}
	}
	return false
}
