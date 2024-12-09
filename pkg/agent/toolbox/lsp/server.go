package lsp

import (
	"context"
	"sync"
)

type LSPServer interface {
	Initialize(pathToProject string) error
	IsInitialized() bool
	Shutdown() error

	HandleDidOpen(ctx context.Context, uri string) error
	HandleDidClose(ctx context.Context, uri string) error
	HandleCompletion(ctx context.Context, params CompletionParams) (*CompletionList, error)
	HandleDocumentSymbols(ctx context.Context, uri string) ([]interface{}, error)
	HandleWorkspaceSymbols(ctx context.Context, query string) ([]interface{}, error)
}

type LSPServerAbstract struct {
	client *Client
	mu     sync.Mutex

	languageId  string
	initialized bool
}

// Add new request types
type WorkspaceSymbolRequest struct {
	Query string `json:"query"`
}

func (s *LSPServerAbstract) IsInitialized() bool {
	return s.initialized
}

func (s *LSPServerAbstract) HandleDidOpen(ctx context.Context, uri string) error {
	if err := s.client.DidOpen(ctx, uri, s.languageId); err != nil {
		return err
	}

	return nil
}

func (s *LSPServerAbstract) HandleDidClose(ctx context.Context, uri string) error {
	if err := s.client.NotifyDidClose(ctx, uri); err != nil {
		return err
	}

	return nil
}

func (s *LSPServerAbstract) HandleCompletion(ctx context.Context, params CompletionParams) (*CompletionList, error) {
	/*
	   	var params CompletionParams
	       if err := c.ShouldBindJSON(&params); err != nil {
	           c.JSON(400, gin.H{"error": err.Error()})
	           return
	       }
	*/
	completions, err := s.client.GetCompletion(
		ctx,
		params.TextDocument.URI,
		params.Position,
		params.Context,
	)
	if err != nil {
		return nil, err
	}

	return completions, nil
}

func (s *LSPServerAbstract) HandleDocumentSymbols(ctx context.Context, uri string) ([]interface{}, error) {
	symbols, err := s.client.GetDocumentSymbols(ctx, uri)
	if err != nil {
		return nil, err
	}

	return symbols, nil
}

func (s *LSPServerAbstract) HandleWorkspaceSymbols(ctx context.Context, query string) ([]interface{}, error) {
	symbols, err := s.client.GetWorkspaceSymbols(ctx, query)
	if err != nil {
		return nil, err
	}

	return symbols, nil
}
