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

type SearchRequest struct {
	Query         string   `json:"query" validate:"required"`
	Path          string   `json:"path"`
	FileTypes     []string `json:"file_types,omitempty"`
	IncludeGlobs  []string `json:"include_globs,omitempty"`
	ExcludeGlobs  []string `json:"exclude_globs,omitempty"`
	CaseSensitive *bool    `json:"case_sensitive,omitempty"`
	Multiline     *bool    `json:"multiline,omitempty"`
	Context       *int     `json:"context,omitempty"`
	CountOnly     *bool    `json:"count_only,omitempty"`
	FilenamesOnly *bool    `json:"filenames_only,omitempty"`
	JSON          *bool    `json:"json,omitempty"`
	MaxResults    *int     `json:"max_results,omitempty"`
	RgArgs        []string `json:"rg_args,omitempty"`
} // @name SearchRequest

type SearchMatch struct {
	File          string   `json:"file" validate:"required"`
	LineNumber    int      `json:"line_number" validate:"required"`
	Column        int      `json:"column" validate:"required"`
	Line          string   `json:"line" validate:"required"`
	Match         string   `json:"match" validate:"required"`
	ContextBefore []string `json:"context_before,omitempty"`
	ContextAfter  []string `json:"context_after,omitempty"`
} // @name SearchMatch

type SearchResults struct {
	Matches      []SearchMatch `json:"matches" validate:"required"`
	TotalMatches int           `json:"total_matches" validate:"required"`
	TotalFiles   int           `json:"total_files" validate:"required"`
	Files        []string      `json:"files,omitempty"`
	RawOutput    string        `json:"raw_output,omitempty"`
} // @name SearchResults
