// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package lsp

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/jsonrpc2"
)

type Client struct {
	conn *jsonrpc2.Conn
}

type InitializeParams struct {
	ProcessID    int                `json:"processId"`
	ClientInfo   ClientInfo         `json:"clientInfo"`
	RootURI      string             `json:"rootUri"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ClientCapabilities struct {
	TextDocument TextDocumentClientCapabilities `json:"textDocument"`
	Workspace    WorkspaceClientCapabilities    `json:"workspace"`
}

type TextDocumentClientCapabilities struct {
	Completion     CompletionClientCapabilities     `json:"completion"`
	DocumentSymbol DocumentSymbolClientCapabilities `json:"documentSymbol"`
}

type CompletionClientCapabilities struct {
	DynamicRegistration bool                       `json:"dynamicRegistration"`
	CompletionItem      CompletionItemCapabilities `json:"completionItem"`
	ContextSupport      bool                       `json:"contextSupport"`
}

type CompletionItemCapabilities struct {
	SnippetSupport          bool     `json:"snippetSupport"`
	CommitCharactersSupport bool     `json:"commitCharactersSupport"`
	DocumentationFormat     []string `json:"documentationFormat"`
	DeprecatedSupport       bool     `json:"deprecatedSupport"`
	PreselectSupport        bool     `json:"preselectSupport"`
}

type DocumentSymbolClientCapabilities struct {
	DynamicRegistration bool           `json:"dynamicRegistration"`
	SymbolKind          SymbolKindInfo `json:"symbolKind"`
}

type SymbolKindInfo struct {
	ValueSet []int `json:"valueSet"`
}

type WorkspaceClientCapabilities struct {
	Symbol WorkspaceSymbolClientCapabilities `json:"symbol"`
}

type WorkspaceSymbolClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration"`
}

type StdioStream struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out io.ReadCloser
}

type Position struct {
	Line      int `json:"line" validate:"required"`
	Character int `json:"character" validate:"required"`
} // @name Position

type Range struct {
	Start Position `json:"start" validate:"required"`
	End   Position `json:"end" validate:"required"`
} // @name Range

type TextDocumentIdentifier struct {
	URI string `json:"uri" validate:"required"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri" validate:"required"`
	Version int    `json:"version" validate:"required"`
} // @name VersionedTextDocumentIdentifier

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument" validate:"required"`
	Position     Position               `json:"position" validate:"required"`
	Context      *CompletionContext     `json:"context,omitempty" validate:"optional"`
} // @name CompletionParams

type CompletionContext struct {
	TriggerKind      int     `json:"triggerKind" validate:"required"`
	TriggerCharacter *string `json:"triggerCharacter,omitempty" validate:"optional"`
} // @name CompletionContext

type CompletionItem struct {
	Label         string      `json:"label" validate:"required"`
	Kind          *int        `json:"kind,omitempty" validate:"optional"`
	Detail        *string     `json:"detail,omitempty" validate:"optional"`
	Documentation interface{} `json:"documentation,omitempty" validate:"optional"`
	SortText      *string     `json:"sortText,omitempty" validate:"optional"`
	FilterText    *string     `json:"filterText,omitempty" validate:"optional"`
	InsertText    *string     `json:"insertText,omitempty" validate:"optional"`
} // @name CompletionItem

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete" validate:"required"`
	Items        []CompletionItem `json:"items" validate:"required"`
} // @name CompletionList

type LspSymbol struct {
	Kind     int         `json:"kind" validate:"required"`
	Location LspLocation `json:"location" validate:"required"`
	Name     string      `json:"name" validate:"required"`
} // @name LspSymbol

type LspLocation struct {
	Range LspRange `json:"range" validate:"required"`
	URI   string   `json:"uri" validate:"required"`
} // @name LspLocation

type LspRange struct {
	End   LspPosition `json:"end" validate:"required"`
	Start LspPosition `json:"start" validate:"required"`
} // @name LspRange

type LspPosition struct {
	Character int `json:"character" validate:"required"`
	Line      int `json:"line" validate:"required"`
} // @name LspPosition

type WorkspaceSymbolParams struct {
	Query string `json:"query" validate:"required"`
} // @name WorkspaceSymbolParams

func (s *StdioStream) Read(p []byte) (n int, err error) {
	return s.out.Read(p)
}

func (s *StdioStream) Write(p []byte) (n int, err error) {
	return s.in.Write(p)
}

func (s *StdioStream) Close() error {
	if err := s.in.Close(); err != nil {
		return err
	}
	return s.out.Close()
}

func NewStdioStream(cmd *exec.Cmd) (*StdioStream, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return &StdioStream{
		cmd: cmd,
		in:  stdin,
		out: stdout,
	}, nil
}

func (c *Client) NotifyDidClose(ctx context.Context, uri string) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}

	return c.conn.Notify(ctx, "textDocument/didClose", params)
}

func (c *Client) GetWorkspaceSymbols(ctx context.Context, query string) ([]LspSymbol, error) {
	params := map[string]interface{}{
		"query": query,
	}

	var symbols []LspSymbol
	err := c.conn.Call(ctx, "workspace/symbol", params, &symbols)
	return symbols, err
}

func (c *Client) GetCompletion(ctx context.Context, uri string, position Position, context *CompletionContext) (*CompletionList, error) {
	params := CompletionParams{
		TextDocument: TextDocumentIdentifier{
			URI: uri,
		},
		Position: position,
		Context:  context,
	}

	var result interface{}
	if err := c.conn.Call(ctx, "textDocument/completion", params, &result); err != nil {
		return nil, err
	}

	// Handle both possible response types: CompletionList or []CompletionItem
	var completionList CompletionList
	switch v := result.(type) {
	case map[string]interface{}:
		// It's a CompletionList
		if items, ok := v["items"].([]interface{}); ok {
			completionItems := make([]CompletionItem, 0, len(items))
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					completionItems = append(completionItems, parseCompletionItem(itemMap))
				}
			}
			completionList.Items = completionItems
			completionList.IsIncomplete = v["isIncomplete"].(bool)
		}
	case []interface{}:
		// It's an array of CompletionItems
		completionItems := make([]CompletionItem, 0, len(v))
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				completionItems = append(completionItems, parseCompletionItem(itemMap))
			}
		}
		completionList.Items = completionItems
	}

	return &completionList, nil
}

func (c *Client) DidOpen(ctx context.Context, uri string, languageId string) error {
	path := strings.TrimPrefix(uri, "file://")

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        uri,
			"languageId": languageId,
			"version":    1,
			"text":       string(content),
		},
	}

	return c.conn.Notify(ctx, "textDocument/didOpen", params)
}

func (c *Client) GetDocumentSymbols(ctx context.Context, uri string) ([]LspSymbol, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}

	var symbols []LspSymbol
	err := c.conn.Call(ctx, "textDocument/documentSymbol", params, &symbols)
	return symbols, err
}

func parseCompletionItem(item map[string]interface{}) CompletionItem {
	ci := CompletionItem{
		Label: item["label"].(string),
	}

	if kind, ok := item["kind"].(float64); ok {
		k := int(kind)
		ci.Kind = &k
	}

	if detail, ok := item["detail"].(string); ok {
		ci.Detail = &detail
	}

	if sortText, ok := item["sortText"].(string); ok {
		ci.SortText = &sortText
	}

	if filterText, ok := item["filterText"].(string); ok {
		ci.FilterText = &filterText
	}

	if insertText, ok := item["insertText"].(string); ok {
		ci.InsertText = &insertText
	}

	ci.Documentation = item["documentation"]

	return ci
}

func (c *Client) Initialize(ctx context.Context, params InitializeParams) error {
	var result interface{}
	if err := c.conn.Call(ctx, "initialize", params, &result); err != nil {
		return err
	}

	return c.conn.Notify(ctx, "initialized", nil)
}

func (c *Client) Shutdown(ctx context.Context) error {
	return c.conn.Notify(ctx, "shutdown", nil)
}
