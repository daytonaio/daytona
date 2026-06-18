// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package internal

import (
	"os"

	"golang.org/x/term"
)

// NoInput is set by the root --no-input persistent flag. When true, the CLI
// must never prompt for input and should fail instead.
var NoInput bool

// Interactive reports whether the CLI may prompt the user for input:
// prompting was not disabled via --no-input and both stdin and stdout are
// terminals.
func Interactive() bool {
	return !NoInput && term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}
