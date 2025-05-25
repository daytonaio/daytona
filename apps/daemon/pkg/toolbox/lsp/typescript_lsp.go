// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package lsp

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/sourcegraph/jsonrpc2"

	log "github.com/sirupsen/logrus"
)

type TypescriptLSPServer struct {
	*LSPServerAbstract
}

func (s *TypescriptLSPServer) Initialize(pathToProject string) error {
	ctx := context.Background()

	cmd := exec.Command("typescript-language-server", "--stdio")

	stream, err := NewStdioStream(cmd)
	if err != nil {
		return fmt.Errorf("failed to create stdio stream: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start LSP server: %w", err)
	}

	handler := jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
		log.Debugf("Received request: %s", req.Method)
		if req.Params != nil {
			log.Debugf("Params: %+v", req.Params)
		}
		return nil, nil
	})

	conn := jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(stream, jsonrpc2.VSCodeObjectCodec{}), handler)

	client := &Client{conn: conn}

	params := InitializeParams{
		ProcessID: os.Getpid(),
		ClientInfo: ClientInfo{
			Name:    "datyona-typescript-lsp-client",
			Version: "0.0.1",
		},
		RootURI: "file://" + pathToProject,
		Capabilities: ClientCapabilities{
			TextDocument: TextDocumentClientCapabilities{
				Completion: CompletionClientCapabilities{
					DynamicRegistration: true,
					CompletionItem: CompletionItemCapabilities{
						SnippetSupport:          true,
						CommitCharactersSupport: true,
						DocumentationFormat:     []string{"markdown", "plaintext"},
						DeprecatedSupport:       true,
						PreselectSupport:        true,
					},
					ContextSupport: true,
				},
				DocumentSymbol: DocumentSymbolClientCapabilities{
					DynamicRegistration: true,
					SymbolKind: SymbolKindInfo{
						ValueSet: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
					},
				},
			},
			Workspace: WorkspaceClientCapabilities{
				Symbol: WorkspaceSymbolClientCapabilities{
					DynamicRegistration: true,
				},
			},
		},
	}

	if err := client.Initialize(ctx, params); err != nil {
		conn.Close()
		killerr := cmd.Process.Kill()
		if killerr != nil {
			return fmt.Errorf("failed to initialize Typescript LSP connection: %w, failed to kill process: %w", err, killerr)
		}
		return fmt.Errorf("failed to initialize Typescript LSP connection: %w", err)
	}

	s.client = client
	s.initialized = true

	return nil
}

func (s *TypescriptLSPServer) Shutdown() error {
	err := s.client.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("failed to shutdown Typescript LSP server: %w", err)
	}
	s.initialized = false
	return nil
}

func NewTypeScriptLSPServer() *TypescriptLSPServer {
	return &TypescriptLSPServer{
		LSPServerAbstract: &LSPServerAbstract{
			languageId: "typescript",
		},
	}
}
