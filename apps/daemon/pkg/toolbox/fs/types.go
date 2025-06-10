// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

type FileInfo struct {
	Name        string `json:"name" validate:"required"`
	Size        int64  `json:"size" validate:"required"`
	Mode        string `json:"mode" validate:"required"`
	ModTime     string `json:"modTime" validate:"required"`
	IsDir       bool   `json:"isDir" validate:"required"`
	Owner       string `json:"owner" validate:"required"`
	Group       string `json:"group" validate:"required"`
	Permissions string `json:"permissions" validate:"required"`
} // @name FileInfo

type ReplaceRequest struct {
	Files    []string `json:"files" validate:"required"`
	Pattern  string   `json:"pattern" validate:"required"`
	NewValue *string  `json:"newValue" validate:"required"`
} // @name ReplaceRequest

type ReplaceResult struct {
	File    string `json:"file"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
} // @name ReplaceResult

type Match struct {
	File    string `json:"file" validate:"required"`
	Line    int    `json:"line" validate:"required"`
	Content string `json:"content" validate:"required"`
} // @name Match

type SearchFilesResponse struct {
	Files []string `json:"files" validate:"required"`
} // @name SearchFilesResponse
