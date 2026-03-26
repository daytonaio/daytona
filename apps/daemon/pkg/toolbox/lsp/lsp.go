// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package lsp

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Start godoc
//
//	@Summary		Start LSP server
//	@Description	Start a Language Server Protocol server for the specified language
//	@Tags			lsp
//	@Accept			json
//	@Produce		json
//	@Param			request	body	LspServerRequest	true	"LSP server request"
//	@Success		200
//	@Router			/lsp/start [post]
//
//	@id				Start
func Start(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LspServerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		service := GetLSPService(logger)
		err := service.Start(req.LanguageId, req.PathToProject)
		if err != nil {
			logger.Error("error starting LSP server", "error", err)
			c.AbortWithError(http.StatusInternalServerError, errors.New("error starting LSP server"))
			return
		}

		c.Status(http.StatusOK)
	}
}

// Stop godoc
//
//	@Summary		Stop LSP server
//	@Description	Stop a Language Server Protocol server
//	@Tags			lsp
//	@Accept			json
//	@Produce		json
//	@Param			request	body	LspServerRequest	true	"LSP server request"
//	@Success		200
//	@Router			/lsp/stop [post]
//
//	@id				Stop
func Stop(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LspServerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		service := GetLSPService(logger)
		err := service.Shutdown(req.LanguageId, req.PathToProject)
		if err != nil {
			logger.Error("error stopping LSP server", "error", err)
			c.AbortWithError(http.StatusInternalServerError, errors.New("error stopping LSP server"))
			return
		}

		c.Status(http.StatusOK)
	}
}

// DidOpen godoc
//
//	@Summary		Notify document opened
//	@Description	Notify the LSP server that a document has been opened
//	@Tags			lsp
//	@Accept			json
//	@Produce		json
//	@Param			request	body	LspDocumentRequest	true	"Document request"
//	@Success		200
//	@Router			/lsp/did-open [post]
//
//	@id				DidOpen
func DidOpen(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LspDocumentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		service := GetLSPService(logger)
		server, err := service.Get(req.LanguageId, req.PathToProject)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if !server.IsInitialized() {
			c.AbortWithError(http.StatusBadRequest, errors.New("server not initialized"))
			return
		}
		err = server.HandleDidOpen(c.Request.Context(), req.Uri)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.Status(http.StatusOK)
	}
}

// DidClose godoc
//
//	@Summary		Notify document closed
//	@Description	Notify the LSP server that a document has been closed
//	@Tags			lsp
//	@Accept			json
//	@Produce		json
//	@Param			request	body	LspDocumentRequest	true	"Document request"
//	@Success		200
//	@Router			/lsp/did-close [post]
//
//	@id				DidClose
func DidClose(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LspDocumentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		service := GetLSPService(logger)
		server, err := service.Get(req.LanguageId, req.PathToProject)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if !server.IsInitialized() {
			c.AbortWithError(http.StatusBadRequest, errors.New("server not initialized"))
			return
		}
		err = server.HandleDidClose(c.Request.Context(), req.Uri)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.Status(http.StatusOK)
	}
}

// Completions godoc
//
//	@Summary		Get code completions
//	@Description	Get code completion suggestions from the LSP server
//	@Tags			lsp
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LspCompletionParams	true	"Completion request"
//	@Success		200		{object}	CompletionList
//	@Router			/lsp/completions [post]
//
//	@id				Completions
func Completions(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LspCompletionParams
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		service := GetLSPService(logger)
		server, err := service.Get(req.LanguageId, req.PathToProject)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if !server.IsInitialized() {
			c.AbortWithError(http.StatusBadRequest, errors.New("server not initialized"))
			return
		}

		textDocument := TextDocumentIdentifier{
			URI: req.Uri,
		}

		completionParams := CompletionParams{
			TextDocument: textDocument,
			Position:     req.Position,
			Context:      req.Context,
		}

		list, err := server.HandleCompletions(c.Request.Context(), completionParams)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.JSON(http.StatusOK, list)
	}
}

// DocumentSymbols godoc
//
//	@Summary		Get document symbols
//	@Description	Get symbols (functions, classes, etc.) from a document
//	@Tags			lsp
//	@Produce		json
//	@Param			languageId		query	string	true	"Language ID (e.g., python, typescript)"
//	@Param			pathToProject	query	string	true	"Path to project"
//	@Param			uri				query	string	true	"Document URI"
//	@Success		200				{array}	LspSymbol
//	@Router			/lsp/document-symbols [get]
//
//	@id				DocumentSymbols
func DocumentSymbols(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		languageId := c.Query("languageId")
		if languageId == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("languageId is required"))
			return
		}

		pathToProject := c.Query("pathToProject")
		if pathToProject == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("pathToProject is required"))
			return
		}

		uri := c.Query("uri")
		if uri == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("uri is required"))
			return
		}

		service := GetLSPService(logger)
		server, err := service.Get(languageId, pathToProject)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if !server.IsInitialized() {
			c.AbortWithError(http.StatusBadRequest, errors.New("server not initialized"))
			return
		}

		symbols, err := server.HandleDocumentSymbols(c.Request.Context(), uri)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.JSON(http.StatusOK, symbols)
	}
}

// WorkspaceSymbols godoc
//
//	@Summary		Get workspace symbols
//	@Description	Search for symbols across the entire workspace
//	@Tags			lsp
//	@Produce		json
//	@Param			query			query	string	true	"Search query"
//	@Param			languageId		query	string	true	"Language ID (e.g., python, typescript)"
//	@Param			pathToProject	query	string	true	"Path to project"
//	@Success		200				{array}	LspSymbol
//	@Router			/lsp/workspacesymbols [get]
//
//	@id				WorkspaceSymbols
func WorkspaceSymbols(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("query")
		if query == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("query is required"))
			return
		}

		languageId := c.Query("languageId")
		if languageId == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("languageId is required"))
			return
		}

		pathToProject := c.Query("pathToProject")
		if pathToProject == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("pathToProject is required"))
			return
		}

		service := GetLSPService(logger)
		server, err := service.Get(languageId, pathToProject)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if !server.IsInitialized() {
			c.AbortWithError(http.StatusBadRequest, errors.New("server not initialized"))
			return
		}

		symbols, err := server.HandleWorkspaceSymbols(c.Request.Context(), query)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.JSON(http.StatusOK, symbols)
	}
}
