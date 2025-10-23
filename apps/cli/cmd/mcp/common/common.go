// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

const (
	CLAUDE   = "claude"
	CURSOR   = "cursor"
	WINDSURF = "windsurf"

	MCP_LOG_FILE_NAME_FORMAT = "daytona-%s-mcp-server.log"
)

var SupportedDaytonaMCPServers = []string{"sandbox", "fs", "git"} // empty string is for daytona code execution MCP

var SupportedAgents = []string{CLAUDE, CURSOR, WINDSURF}
