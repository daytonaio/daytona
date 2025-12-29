// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"encoding/base64"
	"regexp"
	"strings"
)

// ParseShellWrapper detects and transforms Linux shell wrapper commands
// to extract the actual command for Windows execution.
//
// The SDK sends commands in this format:
// sh -c "export VAR=$(echo 'BASE64' | base64 -d); echo 'BASE64_CMD' | base64 -d | sh"
//
// This function extracts the base64-encoded command and env vars,
// returning a Windows-compatible command.
func ParseShellWrapper(command string) (parsedCommand string, envVars map[string]string) {
	envVars = make(map[string]string)

	// Pattern: sh -c "..."
	shPattern := regexp.MustCompile(`^sh\s+-c\s+"(.+)"$`)
	matches := shPattern.FindStringSubmatch(strings.TrimSpace(command))
	if matches == nil {
		return command, envVars
	}

	inner := matches[1]

	// Extract environment variable exports
	// Pattern: export KEY=$(echo 'BASE64' | base64 -d)
	envPattern := regexp.MustCompile(`export\s+(\w+)=\$\(echo\s+'([^']+)'\s*\|\s*base64\s+-d\)`)
	envMatches := envPattern.FindAllStringSubmatch(inner, -1)
	for _, m := range envMatches {
		key := m[1]
		b64Value := m[2]
		decoded, err := base64.StdEncoding.DecodeString(b64Value)
		if err == nil {
			envVars[key] = string(decoded)
		}
	}

	// Extract the main command
	// Pattern: echo 'BASE64' | base64 -d | sh
	cmdPattern := regexp.MustCompile(`echo\s+'([^']+)'\s*\|\s*base64\s+-d\s*\|\s*sh`)
	cmdMatches := cmdPattern.FindStringSubmatch(inner)
	if cmdMatches != nil {
		b64Cmd := cmdMatches[1]
		decoded, err := base64.StdEncoding.DecodeString(b64Cmd)
		if err == nil {
			parsedCommand = string(decoded)
			return parsedCommand, envVars
		}
	}

	// Couldn't parse, return original
	return command, envVars
}

// BuildWindowsCommand creates a PowerShell command with environment variables
func BuildWindowsCommand(command string, envVars map[string]string) string {
	if len(envVars) == 0 {
		return command
	}

	// Build env var assignments for PowerShell
	var envSetters []string
	for key, value := range envVars {
		// Escape double quotes in value
		escaped := strings.ReplaceAll(value, `"`, "`\"")
		envSetters = append(envSetters, "$env:"+key+"=\""+escaped+"\"")
	}

	return strings.Join(envSetters, "; ") + "; " + command
}
