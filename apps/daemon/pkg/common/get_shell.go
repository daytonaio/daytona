// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"os"
	"strings"
	"sync"
)

// shellsFilePath is a package-level var so tests can point resolution at a
// temporary file.
var shellsFilePath = "/etc/shells"

var (
	shellCacheMu sync.Mutex
	cachedShell  string
	shellCached  bool
)

// GetShell returns the preferred shell for the sandbox. The first successful
// resolution (i.e. /etc/shells was readable) is cached for the daemon
// lifetime; a failed read falls back per-call and is never cached, so a later
// call retries the file.
func GetShell() string {
	shellCacheMu.Lock()
	defer shellCacheMu.Unlock()

	if shellCached {
		return cachedShell
	}

	shell, ok := resolveShell(shellsFilePath)
	if ok {
		cachedShell = shell
		shellCached = true
	}

	return shell
}

// resolveShell reads the shells file at path and picks a shell by preference
// order: /usr/bin/zsh > /bin/zsh > /usr/bin/bash > /bin/bash > $SHELL (if
// set) > sh. The boolean reports whether the file was read successfully.
func resolveShell(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return shellFallback(), false
	}

	var sb strings.Builder
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	shells := sb.String()

	for _, preferred := range []string{"/usr/bin/zsh", "/bin/zsh", "/usr/bin/bash", "/bin/bash"} {
		if strings.Contains(shells, preferred) {
			return preferred, true
		}
	}

	return shellFallback(), true
}

func shellFallback() string {
	if shellEnv, shellSet := os.LookupEnv("SHELL"); shellSet {
		return shellEnv
	}

	return "sh"
}
