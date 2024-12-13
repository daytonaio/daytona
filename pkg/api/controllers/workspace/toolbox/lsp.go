// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import "github.com/gin-gonic/gin"

// LspStart			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Start Lsp server
//	@Description	Start Lsp server process inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string				true	"Workspace ID or Name"
//	@Param			projectId	path	string				true	"Project ID"
//	@Param			params		body	LspServerRequest	true	"LspServerRequest"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/lsp/start [post]
//
//	@id				LspStart
func LspStart(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// LspStop			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Stop Lsp server
//	@Description	Stop Lsp server process inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string				true	"Workspace ID or Name"
//	@Param			projectId	path	string				true	"Project ID"
//	@Param			params		body	LspServerRequest	true	"LspServerRequest"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/lsp/stop [post]
//
//	@id				LspStop
func LspStop(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// LspDidOpen			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Call Lsp DidOpen
//	@Description	The document open notification is sent from the client to the server to signal newly opened text documents.
//	@Produce		json
//	@Param			workspaceId	path	string				true	"Workspace ID or Name"
//	@Param			projectId	path	string				true	"Project ID"
//	@Param			params		body	LspDocumentRequest	true	"LspDocumentRequest"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/lsp/did-open [post]
//
//	@id				LspDidOpen
func LspDidOpen(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// LspDidClose			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Call Lsp DidClose
//	@Description	The document close notification is sent from the client to the server when the document got closed in the client.
//	@Produce		json
//	@Param			workspaceId	path	string				true	"Workspace ID or Name"
//	@Param			projectId	path	string				true	"Project ID"
//	@Param			params		body	LspDocumentRequest	true	"LspDocumentRequest"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/lsp/did-close [post]
//
//	@id				LspDidClose
func LspDidClose(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// LspDocumentSymbols			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Call Lsp DocumentSymbols
//	@Description	The document symbol request is sent from the client to the server.
//	@Produce		json
//	@Param			workspaceId		path	string	true	"Workspace ID or Name"
//	@Param			projectId		path	string	true	"Project ID"
//	@Param			languageId		query	string	true	"Language ID"
//	@Param			pathToProject	query	string	true	"Path to project"
//	@Param			uri				query	string	true	"Document Uri"
//	@Success		200				{array}	LspSymbol
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/lsp/document-symbols [get]
//
//	@id				LspDocumentSymbols
func LspDocumentSymbols(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// LspWorkspaceSymbols			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Call Lsp WorkspaceSymbols
//	@Description	The workspace symbol request is sent from the client to the server to list project-wide symbols matching the query string.
//	@Produce		json
//	@Param			workspaceId		path	string	true	"Workspace ID or Name"
//	@Param			projectId		path	string	true	"Project ID"
//	@Param			languageId		query	string	true	"Language ID"
//	@Param			pathToProject	query	string	true	"Path to project"
//	@Param			query			query	string	true	"Symbol Query"
//	@Success		200				{array}	LspSymbol
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/lsp/workspace-symbols [get]
//
//	@id				LspWorkspaceSymbols
func LspWorkspaceSymbols(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// LspCompletions			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Get Lsp Completions
//	@Description	The Completion request is sent from the client to the server to compute completion items at a given cursor position.
//	@Produce		json
//	@Param			workspaceId	path		string				true	"Workspace ID or Name"
//	@Param			projectId	path		string				true	"Project ID"
//	@Param			params		body		LspCompletionParams	true	"LspCompletionParams"
//	@Success		200			{object}	CompletionList
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/lsp/completions [post]
//
//	@id				LspCompletions
func LspCompletions(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}
