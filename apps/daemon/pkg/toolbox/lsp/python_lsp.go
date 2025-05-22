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

type PythonLSPServer struct {
	*LSPServerAbstract
}

func (s *PythonLSPServer) Initialize(pathToProject string) error {
	ctx := context.Background()

	cmd := exec.Command("pylsp")

	stream, err := NewStdioStream(cmd)
	if err != nil {
		return fmt.Errorf("failed to create stdio stream: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Python LSP server: %w", err)
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
			Name:    "datyona-python-lsp-client",
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
			return fmt.Errorf("failed to initialize Python LSP connection: %w, failed to kill process: %w", err, killerr)
		}
		return fmt.Errorf("failed to initialize Python LSP connection: %w", err)
	}

	s.client = client
	s.initialized = true

	return nil
}

func (s *PythonLSPServer) Shutdown() error {
	err := s.client.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("failed to shutdown Python LSP server: %w", err)
	}
	s.initialized = false
	return nil
}

func NewPythonLSPServer() *PythonLSPServer {
	return &PythonLSPServer{
		LSPServerAbstract: &LSPServerAbstract{
			languageId: "python",
		},
	}
}
