// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/daytonaio/daytona/libs/toolbox-api-client-go"
	"github.com/gorilla/websocket"
)

// CodeInterpreterService provides Python code execution capabilities for a sandbox.
//
// CodeInterpreterService enables running Python code in isolated execution contexts
// with support for streaming output, persistent state, and environment variables.
// It uses WebSockets for real-time output streaming. Access through [Sandbox.CodeInterpreter].
//
// Example:
//
//	// Simple code execution
//	channels, err := sandbox.CodeInterpreter.RunCode(ctx, "print('Hello, World!')")
//	if err != nil {
//	    return err
//	}
//
//	// Wait for completion and get result
//	result := <-channels.Done
//	fmt.Println(result.Stdout)
//
//	// With persistent context
//	ctxInfo, _ := sandbox.CodeInterpreter.CreateContext(ctx, nil)
//	contextID := ctxInfo["id"].(string)
//	channels, _ = sandbox.CodeInterpreter.RunCode(ctx, "x = 42",
//	    options.WithCustomContext(contextID),
//	)
//	<-channels.Done
//	channels, _ = sandbox.CodeInterpreter.RunCode(ctx, "print(x)",
//	    options.WithCustomContext(contextID),
//	)
type CodeInterpreterService struct {
	toolboxClient *toolbox.APIClient
	otel          *otelState
}

// NewCodeInterpreterService creates a new CodeInterpreterService.
//
// This is typically called internally by the SDK when creating a [Sandbox].
// Users should access CodeInterpreterService through [Sandbox.CodeInterpreter]
// rather than creating it directly.
func NewCodeInterpreterService(toolboxClient *toolbox.APIClient, otel *otelState) *CodeInterpreterService {
	return &CodeInterpreterService{
		toolboxClient: toolboxClient,
		otel:          otel,
	}
}

// OutputChannels provides channels for streaming execution output.
//
// All channels are closed when execution completes or encounters an error.
// The Done channel always receives exactly one message with the final result.
type OutputChannels struct {
	Stdout <-chan *types.OutputMessage   // Receives stdout messages as they occur
	Stderr <-chan *types.OutputMessage   // Receives stderr messages as they occur
	Errors <-chan *types.ExecutionError  // Receives execution errors
	Done   <-chan *types.ExecutionResult // Receives final result when execution completes
}

// RunCode executes Python code and returns channels for streaming output.
//
// This method establishes a WebSocket connection to execute code asynchronously,
// streaming stdout and stderr as they become available.
//
// Optional parameters can be configured using functional options:
//   - [options.WithCustomContext]: Use a persistent context for state
//   - [options.WithEnv]: Set environment variables
//   - [options.WithInterpreterTimeout]: Set execution timeout
//
// Example:
//
//	// Basic execution
//	channels, err := sandbox.CodeInterpreter.RunCode(ctx, `
//	    for i in range(5):
//	        print(f"Count: {i}")
//	`)
//	if err != nil {
//	    return err
//	}
//
//	// Stream output
//	for msg := range channels.Stdout {
//	    fmt.Print(msg.Text)
//	}
//
//	// Get final result
//	result := <-channels.Done
//	if result.Error != nil {
//	    fmt.Printf("Error: %s\n", result.Error.Value)
//	}
//
//	// With options
//	channels, err := sandbox.CodeInterpreter.RunCode(ctx, "import os; print(os.environ['API_KEY'])",
//	    options.WithEnv(map[string]string{"API_KEY": "secret"}),
//	    options.WithInterpreterTimeout(30*time.Second),
//	)
//
// Returns [OutputChannels] for receiving streamed output, or an error if connection fails.
func (c *CodeInterpreterService) RunCode(ctx context.Context, code string, opts ...func(*options.RunCode)) (*OutputChannels, error) {
	return withInstrumentation(ctx, c.otel, "CodeInterpreter", "RunCode", func(ctx context.Context) (*OutputChannels, error) {
		runOpts := options.Apply(opts...)

		// Extract values from options
		contextID := runOpts.ContextID
		env := runOpts.Env
		timeout := runOpts.Timeout

		// Create channels for output streaming
		stdoutChan := make(chan *types.OutputMessage, 10)
		stderrChan := make(chan *types.OutputMessage, 10)
		errorChan := make(chan *types.ExecutionError, 10)
		doneChan := make(chan *types.ExecutionResult, 1)

		channels := &OutputChannels{
			Stdout: stdoutChan,
			Stderr: stderrChan,
			Errors: errorChan,
			Done:   doneChan,
		}

		// Start goroutine to handle WebSocket communication
		go func() {
			defer close(stdoutChan)
			defer close(stderrChan)
			defer close(errorChan)
			defer close(doneChan)

			// Build WebSocket URL
			baseURL := c.toolboxClient.GetConfig().Servers[0].URL
			wsURL, err := c.buildWebSocketURL(baseURL, "/process/interpreter/execute")
			if err != nil {
				doneChan <- &types.ExecutionResult{
					Error: &types.ExecutionError{
						Name:  "ConnectionError",
						Value: err.Error(),
					},
				}
				return
			}

			// Create WebSocket connection
			headers := c.buildHeaders(c.toolboxClient)
			conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, headers)
			if err != nil {
				doneChan <- &types.ExecutionResult{
					Error: &types.ExecutionError{
						Name:  "ConnectionError",
						Value: fmt.Sprintf("Failed to connect to WebSocket: %v", err),
					},
				}
				return
			}
			defer conn.Close()

			// Send execute request
			executeReq := map[string]interface{}{
				"code": code,
			}
			if contextID != "" {
				executeReq["contextId"] = contextID
			}
			if env != nil {
				executeReq["envs"] = env
			}
			if timeout != nil {
				timeoutInt64 := int64(timeout.Seconds())
				executeReq["timeout"] = &timeoutInt64
			}

			if err := conn.WriteJSON(executeReq); err != nil {
				doneChan <- &types.ExecutionResult{
					Error: &types.ExecutionError{
						Name:  "ConnectionError",
						Value: fmt.Sprintf("Failed to send execute request: %v", err),
					},
				}
				return
			}

			// Read output messages from WebSocket
			result := &types.ExecutionResult{}
			var stdout, stderr strings.Builder

			for {
				var msg types.OutputMessage
				err := conn.ReadJSON(&msg)
				if err != nil {
					// Check if it's a normal WebSocket close
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						break
					}
					// Check for custom timeout close code (4001)
					if websocket.IsCloseError(err, 4001) {
						result.Error = &types.ExecutionError{
							Name:  "TimeoutError",
							Value: "Execution timed out",
						}
						break
					}
					result.Error = &types.ExecutionError{
						Name:  "ConnectionError",
						Value: fmt.Sprintf("Failed to read message: %v", err),
					}
					break
				}

				// Send to channels based on message type
				switch msg.Type {
				case "stdout":
					stdoutChan <- &msg
					stdout.WriteString(msg.Text)

				case "stderr":
					stderrChan <- &msg
					stderr.WriteString(msg.Text)

				case "error":
					execError := &types.ExecutionError{
						Name:  msg.Name,
						Value: msg.Value,
					}
					if msg.Traceback != "" {
						execError.Traceback = &msg.Traceback
					}
					errorChan <- execError
					result.Error = execError
				}
			}

			result.Stdout = stdout.String()
			result.Stderr = stderr.String()

			doneChan <- result
		}()

		return channels, nil
	})
}

// buildWebSocketURL converts an HTTP(S) URL to a WebSocket URL
func (c *CodeInterpreterService) buildWebSocketURL(baseURL, path string) (string, error) {
	u, err := url.Parse(baseURL + path)
	if err != nil {
		return "", err
	}

	// Convert http/https to ws/wss
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		return "", errors.NewDaytonaError(fmt.Sprintf("Unsupported scheme: %s", u.Scheme), 0, nil)
	}

	return u.String(), nil
}

// buildHeaders builds HTTP headers for the WebSocket connection
func (c *CodeInterpreterService) buildHeaders(toolboxClient *toolbox.APIClient) map[string][]string {
	headers := make(map[string][]string)

	// Copy default headers from toolbox client config
	for key, value := range toolboxClient.GetConfig().DefaultHeader {
		headers[key] = []string{value}
	}

	return headers
}

// CreateContext creates an isolated execution context for persistent state.
//
// Contexts allow you to maintain state (variables, imports, etc.) across
// multiple code executions. Without a context, each RunCode call starts fresh.
//
// Parameters:
//   - cwd: Optional working directory for the context
//
// Example:
//
//	// Create a context
//	ctxInfo, err := sandbox.CodeInterpreter.CreateContext(ctx, nil)
//	if err != nil {
//	    return err
//	}
//	contextID := ctxInfo["id"].(string)
//
//	// Use the context to maintain state
//	sandbox.CodeInterpreter.RunCode(ctx, "x = 42", options.WithCustomContext(contextID))
//	sandbox.CodeInterpreter.RunCode(ctx, "print(x)", options.WithCustomContext(contextID)) // prints 42
//
//	// Clean up when done
//	sandbox.CodeInterpreter.DeleteContext(ctx, contextID)
//
// Returns context information including "id", "cwd", "language", "active", and "createdAt".
func (c *CodeInterpreterService) CreateContext(ctx context.Context, cwd *string) (map[string]any, error) {
	return withInstrumentation(ctx, c.otel, "CodeInterpreter", "CreateContext", func(ctx context.Context) (map[string]any, error) {
		req := toolbox.NewCreateContextRequest()
		if cwd != nil {
			req.SetCwd(*cwd)
		}

		context, httpResp, err := c.toolboxClient.InterpreterAPI.CreateInterpreterContext(ctx).Request(*req).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to map for backward compatibility
		return map[string]any{
			"id":        context.GetId(),
			"cwd":       context.GetCwd(),
			"language":  context.GetLanguage(),
			"active":    context.GetActive(),
			"createdAt": context.GetCreatedAt(),
		}, nil
	})
}

// ListContexts returns all active execution contexts.
//
// Example:
//
//	contexts, err := sandbox.CodeInterpreter.ListContexts(ctx)
//	if err != nil {
//	    return err
//	}
//	for _, ctx := range contexts {
//	    fmt.Printf("Context %s (language: %s)\n", ctx["id"], ctx["language"])
//	}
//
// Returns a slice of context information maps.
func (c *CodeInterpreterService) ListContexts(ctx context.Context) ([]map[string]any, error) {
	return withInstrumentation(ctx, c.otel, "CodeInterpreter", "ListContexts", func(ctx context.Context) ([]map[string]any, error) {
		resp, httpResp, err := c.toolboxClient.InterpreterAPI.ListInterpreterContexts(ctx).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to map array for backward compatibility
		contexts := resp.GetContexts()
		result := make([]map[string]any, len(contexts))
		for i, context := range contexts {
			result[i] = map[string]any{
				"id":        context.GetId(),
				"cwd":       context.GetCwd(),
				"language":  context.GetLanguage(),
				"active":    context.GetActive(),
				"createdAt": context.GetCreatedAt(),
			}
		}

		return result, nil
	})
}

// DeleteContext removes an execution context and releases its resources.
//
// Parameters:
//   - contextID: The context identifier to delete
//
// Example:
//
//	err := sandbox.CodeInterpreter.DeleteContext(ctx, contextID)
//
// Returns an error if the context doesn't exist or deletion fails.
func (c *CodeInterpreterService) DeleteContext(ctx context.Context, contextID string) error {
	return withInstrumentationVoid(ctx, c.otel, "CodeInterpreter", "DeleteContext", func(ctx context.Context) error {
		_, httpResp, err := c.toolboxClient.InterpreterAPI.DeleteInterpreterContext(ctx, contextID).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}
