// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package lsp

type LspServerRequest struct {
	LanguageId    string `json:"languageId" validate:"required"`
	PathToProject string `json:"pathToProject" validate:"required"`
} // @name LspServerRequest

type LspDocumentRequest struct {
	LanguageId    string `json:"languageId" validate:"required"`
	PathToProject string `json:"pathToProject" validate:"required"`
	Uri           string `json:"uri" validate:"required"`
} // @name LspDocumentRequest

type LspCompletionParams struct {
	LanguageId    string             `json:"languageId" validate:"required"`
	PathToProject string             `json:"pathToProject" validate:"required"`
	Uri           string             `json:"uri" validate:"required"`
	Position      Position           `json:"position" validate:"required"`
	Context       *CompletionContext `json:"context,omitempty" validate:"optional"`
} // @name LspCompletionParams
