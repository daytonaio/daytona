// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package security

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const MaxCommandLength = 4096

var forbiddenTokens = []string{"&&", "||", ";", "|", "`", "$(", ">", "<", "\n", "\r", "\x00"}

var allowedExecutables = map[string]struct{}{
	"awk": {}, "bash": {}, "cat": {}, "chmod": {}, "chown": {}, "cp": {}, "curl": {}, "cut": {},
	"date": {}, "df": {}, "du": {}, "echo": {}, "env": {}, "find": {}, "git": {}, "go": {},
	"grep": {}, "gunzip": {}, "gzip": {}, "head": {}, "id": {}, "ls": {}, "make": {}, "mkdir": {},
	"mv": {}, "node": {}, "nohup": {}, "npm": {}, "pip": {}, "pip3": {}, "pnpm": {}, "printenv": {},
	"pwd": {}, "python": {}, "python3": {}, "rm": {}, "sed": {}, "sh": {}, "sleep": {}, "sort": {},
	"stat": {}, "tail": {}, "tar": {}, "touch": {}, "tr": {}, "uname": {}, "uniq": {}, "unzip": {},
	"wc": {}, "wget": {}, "which": {}, "whoami": {}, "yarn": {}, "zip": {}, "zsh": {},
}

func UnsafeCommandChecksDisabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("DAYTONA_ALLOW_UNSAFE_COMMANDS")))
	return value == "1" || value == "true" || value == "yes"
}

func ValidateCommand(command string) error {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return fmt.Errorf("empty command")
	}

	if len(trimmed) > MaxCommandLength {
		return fmt.Errorf("command exceeds %d characters", MaxCommandLength)
	}

	for _, token := range forbiddenTokens {
		if strings.Contains(trimmed, token) {
			return fmt.Errorf("command contains forbidden token %q", token)
		}
	}

	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	executable := filepath.Base(parts[0])
	if _, ok := allowedExecutables[executable]; !ok {
		return fmt.Errorf("executable %q is not allowlisted", executable)
	}

	return nil
}

func ValidateCwd(cwd string) error {
	if strings.Contains(cwd, "\x00") || strings.Contains(cwd, "\n") || strings.Contains(cwd, "\r") {
		return fmt.Errorf("invalid cwd: contains control characters")
	}

	if strings.TrimSpace(cwd) == "" {
		return fmt.Errorf("invalid cwd: empty path")
	}

	return nil
}

func AuditCommandDecision(logger *slog.Logger, source string, command string, allowed bool, reason string) {
	hash := sha256.Sum256([]byte(command))
	commandHash := fmt.Sprintf("%x", hash[:8])

	cmd := strings.TrimSpace(command)
	parts := strings.Fields(cmd)
	executable := ""
	if len(parts) > 0 {
		executable = filepath.Base(parts[0])
	}

	if allowed {
		logger.Info("command policy accepted",
			"source", source,
			"command_hash", commandHash,
			"executable", executable,
		)
		return
	}

	logger.Warn("command policy denied",
		"source", source,
		"command_hash", commandHash,
		"executable", executable,
		"reason", reason,
	)
}
