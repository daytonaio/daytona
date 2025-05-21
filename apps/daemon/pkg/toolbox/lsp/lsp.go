// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package lsp

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func Start(c *gin.Context) {
	var req LspServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	service := GetLSPService()
	err := service.Start(req.LanguageId, req.PathToProject)
	if err != nil {
		log.Error(err)
		c.AbortWithError(http.StatusInternalServerError, errors.New("error starting LSP server"))
		return
	}

	c.Status(http.StatusOK)
}

func Stop(c *gin.Context) {
	var req LspServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	service := GetLSPService()
	err := service.Shutdown(req.LanguageId, req.PathToProject)
	if err != nil {
		log.Error(err)
		c.AbortWithError(http.StatusInternalServerError, errors.New("error stopping LSP server"))
		return
	}

	c.Status(http.StatusOK)
}

func DidOpen(c *gin.Context) {
	var req LspDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	service := GetLSPService()
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

func DidClose(c *gin.Context) {
	var req LspDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	service := GetLSPService()
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

func Completions(c *gin.Context) {
	var req LspCompletionParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	service := GetLSPService()
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

func DocumentSymbols(c *gin.Context) {
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

	service := GetLSPService()
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

func WorkspaceSymbols(c *gin.Context) {
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

	service := GetLSPService()
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
