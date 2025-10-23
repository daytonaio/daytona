package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func GetRunCodeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "run_code",
		Title:       "Run Code",
		Description: "Run code in the Daytona sandbox.",
	}
}

func HandleRunCode(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{}, nil
}
