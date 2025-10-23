package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func GetShellTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "shell",
		Title:       "Shell",
		Description: "Execute shell commands in the Daytona sandbox.",
	}
}

func HandleShell(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{}, nil
}
