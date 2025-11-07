// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daytona

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/daytonaio/mcp/internal/constants"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type CodeRunInput struct {
	SandboxId *string       `json:"sandboxId,omitempty" jsonschema:"ID of the sandbox to run the code in. Don't provide this if not explicitly instructed from user. If not provided, a new sandbox will be created."`
	Language  string        `json:"language,omitempty" jsonschema:"Language to run the code in. Supported languages are: python, typescript, javascript. If not provided, try and guess the language from the code."`
	Code      string        `json:"code" jsonschema:"Code to run."`
	Params    CodeRunParams `json:"params,omitempty" jsonschema:"Parameters for the code run."`
	Timeout   *int          `json:"timeout,omitempty" jsonschema:"Maximum time in seconds to wait for the code to complete. If not provided, the default timeout 0 (meaning indefinitely) will be used."`
}

type CodeRunParams struct {
	Argv []string          `json:"argv,omitempty" jsonschema:"Command line arguments."`
	Env  map[string]string `json:"env,omitempty" jsonschema:"Environment variables."`
}

type CodeRunOutput struct {
	ExitCode             *int                   `json:"exitCode,omitempty" jsonschema:"Exit code of the code run."`
	Result               *string                `json:"result,omitempty" jsonschema:"Result of the code run."`
	Artifacts            *ExecutionArtifacts    `json:"artifacts,omitempty" jsonschema:"Artifacts of the code run."`
	AdditionalProperties map[string]interface{} `json:"additionalProperties,omitempty" jsonschema:"Additional properties."`
}

func (s *DaytonaMCPServer) getCodeRunTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "code_run",
		Title:       "Code Run",
		Description: "Run code in the Daytona sandbox using the appropriate language runtime.",
	}
}

func (s *DaytonaMCPServer) handleCodeRunTool(ctx context.Context, request *mcp.CallToolRequest, input *CodeRunInput) (*mcp.CallToolResult, *CodeRunOutput, error) {
	if input.Language == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("language is required")
	}

	if input.Code == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("code is required")
	}

	// Get the run command based on language
	runCommand := getRunCommand(input.Language, input.Code, input.Params)
	if runCommand == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("unsupported language: %s", input.Language)
	}

	// Use executeCommand to run the code
	executeInput := &ExecuteCommandInput{
		SandboxId: input.SandboxId,
		Command:   runCommand,
		Env:       input.Params.Env,
		Timeout:   input.Timeout,
	}

	result, output, err := s.handleExecuteCommand(ctx, request, executeInput)
	if err != nil {
		return result, nil, err
	}

	// Convert ExecuteCommandOutput to ExecuteResponse
	var exitCode *int
	var resultStr *string
	var artifacts *ExecutionArtifacts

	if output != nil {
		exitCode = output.ExitCode
		resultStr = output.Result
		artifacts = output.Artifacts
	}

	executeResponse := &CodeRunOutput{
		ExitCode:  exitCode,
		Result:    resultStr,
		Artifacts: artifacts,
	}

	return result, executeResponse, nil
}

// getRunCommand returns the command to execute code for the given language
func getRunCommand(language, code string, params CodeRunParams) string {
	switch language {
	case "python":
		return getPythonRunCommand(code, params)
	case "typescript":
		return getTypescriptRunCommand(code, params)
	case "javascript":
		return getJavascriptRunCommand(code, params)
	default:
		return ""
	}
}

// getJavascriptRunCommand generates the command to run JavaScript code
func getJavascriptRunCommand(code string, params CodeRunParams) string {
	base64Code := base64.StdEncoding.EncodeToString([]byte(code))
	argv := ""
	if len(params.Argv) > 0 {
		argv = strings.Join(params.Argv, " ")
	}
	return fmt.Sprintf("sh -c 'echo %s | base64 --decode | node -e \"$(cat)\" %s 2>&1 | grep -vE \"npm notice\"'", base64Code, argv)
}

// getTypescriptRunCommand generates the command to run TypeScript code
func getTypescriptRunCommand(code string, params CodeRunParams) string {
	base64Code := base64.StdEncoding.EncodeToString([]byte(code))
	argv := ""
	if len(params.Argv) > 0 {
		argv = strings.Join(params.Argv, " ")
	}
	// Note: The escaped quotes in the TypeScript command need careful handling
	return fmt.Sprintf("sh -c 'echo %s | base64 --decode | npx ts-node -O \"{\\\"module\\\":\\\"CommonJS\\\"}\" -e \"$(cat)\" x %s 2>&1 | grep -vE \"npm notice\"'", base64Code, argv)
}

// getPythonRunCommand generates the command to run Python code
func getPythonRunCommand(code string, params CodeRunParams) string {
	// Encode the provided code in base64
	base64Code := base64.StdEncoding.EncodeToString([]byte(code))

	// Override plt.show() method if matplotlib is imported
	if isMatplotlibImported(code) {
		// Replace {encoded_code} with the actual base64-encoded code
		codeWrapper := strings.ReplaceAll(constants.PYTHON_CODE_WRAPPER, "{encoded_code}", base64Code)
		base64Code = base64.StdEncoding.EncodeToString([]byte(codeWrapper))
	}

	// Build command-line arguments string
	argv := ""
	if len(params.Argv) > 0 {
		argv = strings.Join(params.Argv, " ")
	}

	// Execute the bootstrapper code directly
	// Use -u flag to ensure unbuffered output for real-time error reporting
	return fmt.Sprintf("sh -c 'python3 -u -c \"exec(__import__(\\\"base64\\\").b64decode(\\\"%s\\\").decode())\" %s'", base64Code, argv)
}

// isMatplotlibImported checks if matplotlib is imported in the given Python code string
func isMatplotlibImported(codeString string) bool {
	// Regex patterns for different import styles
	patterns := []*regexp.Regexp{
		// Standard imports
		regexp.MustCompile(`(?m)^[^#]*import\s+matplotlib`),
		regexp.MustCompile(`(?m)^[^#]*from\s+matplotlib`),

		// Dynamic imports
		regexp.MustCompile(`(?m)^[^#]*__import__\s*\(\s*['"]matplotlib['"]`),
		regexp.MustCompile(`(?m)^[^#]*importlib\.import_module\s*\(\s*['"]matplotlib['"]`),

		// Other dynamic loading patterns
		regexp.MustCompile(`(?m)^[^#]*loader\.load_module\s*\(\s*['"]matplotlib['"]`),
		regexp.MustCompile(`(?m)^[^#]*sys\.modules\[['"]matplotlib['"]\]`),
	}

	// Check each pattern
	for _, pattern := range patterns {
		if pattern.MatchString(codeString) {
			return true
		}
	}

	return false
}
