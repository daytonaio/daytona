// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package lsp

import (
	"errors"

	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func Start(c *gin.Context) {
	var req LspServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	service := GetLSPService()
	err := service.Start(req.LanguageId, req.PathToProject)
	if err != nil {
		log.Error(err)
		c.AbortWithError(500, errors.New("error starting LSP server"))
		return
	}

	c.Status(200)
}

func Stop(c *gin.Context) {
	var req LspServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	service := GetLSPService()
	err := service.Shutdown(req.LanguageId, req.PathToProject)
	if err != nil {
		log.Error(err)
		c.AbortWithError(500, errors.New("error stopping LSP server"))
		return
	}

	c.Status(200)
}

func DidOpen(c *gin.Context) {
	var req LspDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	service := GetLSPService()
	server, err := service.Get(req.LanguageId, req.PathToProject)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	if !server.IsInitialized() {
		c.AbortWithError(400, errors.New("server not initialized"))
		return
	}
	err = server.HandleDidOpen(c.Request.Context(), req.Uri)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.Status(200)
}

func DidClose(c *gin.Context) {
	var req LspDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	service := GetLSPService()
	server, err := service.Get(req.LanguageId, req.PathToProject)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	if !server.IsInitialized() {
		c.AbortWithError(400, errors.New("server not initialized"))
		return
	}
	err = server.HandleDidClose(c.Request.Context(), req.Uri)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.Status(200)
}

func Completions(c *gin.Context) {
	var req LspCompletionParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	service := GetLSPService()
	server, err := service.Get(req.LanguageId, req.PathToProject)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	if !server.IsInitialized() {
		c.AbortWithError(400, errors.New("server not initialized"))
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
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, list)
}

func DocumentSymbols(c *gin.Context) {
	languageId := c.Query("languageId")
	if languageId == "" {
		c.AbortWithError(400, errors.New("languageId is required"))
		return
	}

	pathToProject := c.Query("pathToProject")
	if languageId == "" {
		c.AbortWithError(400, errors.New("pathToProject is required"))
		return
	}

	uri := c.Query("uri")
	if uri == "" {
		c.AbortWithError(400, errors.New("uri is required"))
		return
	}

	service := GetLSPService()
	server, err := service.Get(languageId, pathToProject)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	if !server.IsInitialized() {
		c.AbortWithError(400, errors.New("server not initialized"))
		return
	}

	symbols, err := server.HandleDocumentSymbols(c.Request.Context(), uri)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, symbols)
}

func WorkspaceSymbols(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.AbortWithError(400, errors.New("query is required"))
		return
	}

	languageId := c.Query("languageId")
	if languageId == "" {
		c.AbortWithError(400, errors.New("languageId is required"))
		return
	}

	pathToProject := c.Query("pathToProject")
	if languageId == "" {
		c.AbortWithError(400, errors.New("pathToProject is required"))
		return
	}

	service := GetLSPService()
	server, err := service.Get(languageId, pathToProject)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	if !server.IsInitialized() {
		c.AbortWithError(400, errors.New("server not initialized"))
		return
	}

	symbols, err := server.HandleWorkspaceSymbols(c.Request.Context(), query)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, symbols)
}
