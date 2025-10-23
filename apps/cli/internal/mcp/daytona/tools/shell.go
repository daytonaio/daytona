package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ShellInput struct {
	Command string `json:"command" jsonchema:"Command to execute."`
	Timeout int    `json:"timeout" jsonchema:"Maximum time in seconds to wait for the command to complete. If not provided, the default timeout 0 (meaning indefinitely) will be used."`
}

type ShellOutput struct {
	Stdout   string `json:"stdout" jsonchema:"Standard output of the command."`
	Stderr   string `json:"stderr" jsonchema:"Standard error output of the command."`
	ExitCode int    `json:"exitCode" jsonchema:"Exit code of the command."`
}

func GetShellTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "shell",
		Title:       "Shell",
		Description: "Execute shell commands in the Daytona sandbox.",
	}
}

func HandleShell(ctx context.Context, request *mcp.CallToolRequest, input *ShellInput) (*mcp.CallToolResult, *ShellOutput, error) {
	return &mcp.CallToolResult{}, nil, nil
}
