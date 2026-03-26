// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"strings"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

// LspServerService provides Language Server Protocol (LSP) operations for a sandbox.
//
// LspServerService enables IDE-like features such as code completion, symbol search,
// and document analysis through LSP. The service manages a language server instance
// for a specific language and project path. Access through [Sandbox.Lsp].
//
// Example:
//
//	// Get LSP service for Python
//	lsp := sandbox.Lsp(types.LspLanguageIDPython, "/home/user/project")
//
//	// Start the language server
//	if err := lsp.Start(ctx); err != nil {
//	    return err
//	}
//	defer lsp.Stop(ctx)
//
//	// Open a file for analysis
//	if err := lsp.DidOpen(ctx, "/home/user/project/main.py"); err != nil {
//	    return err
//	}
//
//	// Get code completions
//	completions, err := lsp.Completions(ctx, "/home/user/project/main.py",
//	    types.Position{Line: 10, Character: 5})
type LspServerService struct {
	toolboxClient *toolbox.APIClient
	languageID    types.LspLanguageID
	projectPath   string
	otel          *otelState
}

// NewLspServerService creates a new LspServerService.
//
// This is typically called internally by the SDK through [Sandbox.Lsp].
// Users should access LspServerService through [Sandbox.Lsp] rather than
// creating it directly.
//
// Parameters:
//   - toolboxClient: The toolbox API client
//   - languageID: The language identifier (e.g., [types.LspLanguageIDPython])
//   - projectPath: The root path of the project for LSP analysis
func NewLspServerService(toolboxClient *toolbox.APIClient, languageID types.LspLanguageID, projectPath string, otel *otelState) *LspServerService {
	return &LspServerService{
		toolboxClient: toolboxClient,
		languageID:    languageID,
		projectPath:   projectPath,
		otel:          otel,
	}
}

// Start initializes and starts the language server.
//
// The language server must be started before using other LSP operations.
// Call [LspServerService.Stop] when finished to release resources.
//
// Example:
//
//	if err := lsp.Start(ctx); err != nil {
//	    return err
//	}
//	defer lsp.Stop(ctx)
//
// Returns an error if the server fails to start.
func (l *LspServerService) Start(ctx context.Context) error {
	return withInstrumentationVoid(ctx, l.otel, "LspServer", "Start", func(ctx context.Context) error {
		req := toolbox.NewLspServerRequest(string(l.languageID), l.projectPath)
		httpResp, err := l.toolboxClient.LspAPI.Start(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// Stop shuts down the language server and releases resources.
//
// Example:
//
//	err := lsp.Stop(ctx)
//
// Returns an error if the server fails to stop gracefully.
func (l *LspServerService) Stop(ctx context.Context) error {
	return withInstrumentationVoid(ctx, l.otel, "LspServer", "Stop", func(ctx context.Context) error {
		req := toolbox.NewLspServerRequest(string(l.languageID), l.projectPath)
		httpResp, err := l.toolboxClient.LspAPI.Stop(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// DidOpen notifies the language server that a file was opened.
//
// This should be called before requesting completions or symbols for a file.
// The path is automatically converted to a file:// URI if needed.
//
// Parameters:
//   - path: Absolute path to the file
//
// Example:
//
//	err := lsp.DidOpen(ctx, "/home/user/project/main.py")
//
// Returns an error if the notification fails.
func (l *LspServerService) DidOpen(ctx context.Context, path string) error {
	return withInstrumentationVoid(ctx, l.otel, "LspServer", "DidOpen", func(ctx context.Context) error {
		uri := path
		if !strings.HasPrefix(uri, "file://") {
			uri = "file://" + uri
		}

		req := toolbox.NewLspDocumentRequest(string(l.languageID), l.projectPath, uri)
		httpResp, err := l.toolboxClient.LspAPI.DidOpen(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// DidClose notifies the language server that a file was closed.
//
// Call this when you're done working with a file to allow the server
// to release resources associated with it.
//
// Parameters:
//   - path: Absolute path to the file
//
// Example:
//
//	err := lsp.DidClose(ctx, "/home/user/project/main.py")
//
// Returns an error if the notification fails.
func (l *LspServerService) DidClose(ctx context.Context, path string) error {
	return withInstrumentationVoid(ctx, l.otel, "LspServer", "DidClose", func(ctx context.Context) error {
		uri := path
		if !strings.HasPrefix(uri, "file://") {
			uri = "file://" + uri
		}

		req := toolbox.NewLspDocumentRequest(string(l.languageID), l.projectPath, uri)
		httpResp, err := l.toolboxClient.LspAPI.DidClose(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// DocumentSymbols returns all symbols (functions, classes, variables) in a document.
//
// Parameters:
//   - path: Absolute path to the file
//
// Example:
//
//	symbols, err := lsp.DocumentSymbols(ctx, "/home/user/project/main.py")
//	if err != nil {
//	    return err
//	}
//	for _, sym := range symbols {
//	    fmt.Printf("Symbol: %v\n", sym)
//	}
//
// Returns a slice of symbol information or an error.
func (l *LspServerService) DocumentSymbols(ctx context.Context, path string) ([]any, error) {
	return withInstrumentation(ctx, l.otel, "LspServer", "DocumentSymbols", func(ctx context.Context) ([]any, error) {
		uri := path
		if !strings.HasPrefix(uri, "file://") {
			uri = "file://" + uri
		}

		symbols, httpResp, err := l.toolboxClient.LspAPI.DocumentSymbols(ctx).
			LanguageId(string(l.languageID)).
			PathToProject(l.projectPath).
			Uri(uri).
			Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to []any for backward compatibility
		result := make([]any, len(symbols))
		for i, symbol := range symbols {
			result[i] = symbol
		}
		return result, nil
	})
}

// SandboxSymbols searches for symbols across the entire workspace.
//
// Use this to find symbols (functions, classes, etc.) by name across all files
// in the project.
//
// Parameters:
//   - query: Search query to match symbol names
//
// Example:
//
//	symbols, err := lsp.SandboxSymbols(ctx, "MyClass")
//	if err != nil {
//	    return err
//	}
//	for _, sym := range symbols {
//	    fmt.Printf("Found: %v\n", sym)
//	}
//
// Returns a slice of matching symbols or an error.
func (l *LspServerService) SandboxSymbols(ctx context.Context, query string) ([]any, error) {
	return withInstrumentation(ctx, l.otel, "LspServer", "SandboxSymbols", func(ctx context.Context) ([]any, error) {
		symbols, httpResp, err := l.toolboxClient.LspAPI.WorkspaceSymbols(ctx).
			LanguageId(string(l.languageID)).
			PathToProject(l.projectPath).
			Query(query).
			Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to []any for backward compatibility
		result := make([]any, len(symbols))
		for i, symbol := range symbols {
			result[i] = symbol
		}
		return result, nil
	})
}

// Completions returns code completion suggestions at a position.
//
// The file should be opened with [LspServerService.DidOpen] before requesting completions.
//
// Parameters:
//   - path: Absolute path to the file
//   - position: Cursor position (line and character, 0-indexed)
//
// Example:
//
//	lsp.DidOpen(ctx, "/home/user/project/main.py")
//	completions, err := lsp.Completions(ctx, "/home/user/project/main.py",
//	    types.Position{Line: 10, Character: 5})
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Completions: %v\n", completions)
//
// Returns completion items or an error.
func (l *LspServerService) Completions(ctx context.Context, path string, position types.Position) (any, error) {
	return withInstrumentation(ctx, l.otel, "LspServer", "Completions", func(ctx context.Context) (any, error) {
		uri := path
		if !strings.HasPrefix(uri, "file://") {
			uri = "file://" + uri
		}

		// Create LSP position
		lspPos := toolbox.NewLspPosition(int32(position.Line), int32(position.Character))

		req := toolbox.NewLspCompletionParams(string(l.languageID), l.projectPath, *lspPos, uri)

		completions, httpResp, err := l.toolboxClient.LspAPI.Completions(ctx).Request(*req).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Return as any for backward compatibility
		return completions, nil
	})
}
