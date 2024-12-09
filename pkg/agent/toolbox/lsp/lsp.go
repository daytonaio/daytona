package lsp

import (
	"github.com/gin-gonic/gin"
)

type LspServerRequest struct {
	LanguageId string `json:"languageId" binding:"required"`
} // @name LspServerRequest

type LspDocumentRequest struct {
	LanguageId string `json:"languageId" binding:"required"`
	Uri        string `json:"uri" binding:"required"`
} // @name LspDocumentRequest

type LspCompletionParams struct {
	LanguageId   string                 `json:"languageId" binding:"required"`
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      *CompletionContext     `json:"context,omitempty"`
} // @name LspCompletionParams

func Start(c *gin.Context) {
	var req LspServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	service := GetLSPService()
	err := service.Start(req.LanguageId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "LSP server started"})
}

func Stop(c *gin.Context) {
	var req LspServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	service := GetLSPService()
	err := service.Shutdown(req.LanguageId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "LSP server stopped"})
}

func DidOpen(c *gin.Context) {
	var req LspDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	service := GetLSPService()
	server, err := service.Get(req.LanguageId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if !server.IsInitialized() {
		c.JSON(400, gin.H{"error": "server not initialized"})
		return
	}
	err = server.HandleDidOpen(c.Request.Context(), req.Uri)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "didOpen sent"})
}

func DidClose(c *gin.Context) {
	var req LspDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	service := GetLSPService()
	server, err := service.Get(req.LanguageId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if !server.IsInitialized() {
		c.JSON(400, gin.H{"error": "server not initialized"})
		return
	}
	err = server.HandleDidClose(c.Request.Context(), req.Uri)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "didClose sent"})
}

func Completion(c *gin.Context) {
	var req LspCompletionParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	service := GetLSPService()
	server, err := service.Get(req.LanguageId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if !server.IsInitialized() {
		c.JSON(400, gin.H{"error": "server not initialized"})
		return
	}

	completionParams := CompletionParams{
		TextDocument: req.TextDocument,
		Position:     req.Position,
		Context:      req.Context,
	}

	list, err := server.HandleCompletion(c.Request.Context(), completionParams)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, list)
}

func DocumentSymbols(c *gin.Context) {
	languageId := c.Query("languageId")
	if languageId == "" {
		c.JSON(400, gin.H{"error": "languageId is required"})
		return
	}

	uri := c.Query("uri")
	if uri == "" {
		c.JSON(400, gin.H{"error": "uri is required"})
		return
	}

	service := GetLSPService()
	server, err := service.Get(languageId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if !server.IsInitialized() {
		c.JSON(400, gin.H{"error": "server not initialized"})
		return
	}

	symbols, err := server.HandleDocumentSymbols(c.Request.Context(), uri)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, symbols)
}

func WorkspaceSymbols(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(400, gin.H{"error": "query is required"})
		return
	}

	languageId := c.Query("languageId")
	if languageId == "" {
		c.JSON(400, gin.H{"error": "languageId is required"})
		return
	}

	service := GetLSPService()
	server, err := service.Get(languageId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if !server.IsInitialized() {
		c.JSON(400, gin.H{"error": "server not initialized"})
		return
	}

	symbols, err := server.HandleWorkspaceSymbols(c.Request.Context(), query)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, symbols)
}
