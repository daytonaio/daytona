// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

type FileInfo struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Mode        string `json:"mode"`
	ModTime     string `json:"modTime"`
	IsDir       bool   `json:"isDir"`
	Owner       string `json:"owner"`
	Group       string `json:"group"`
	Permissions string `json:"permissions"`
} // @name FileInfo

type ReplaceRequest struct {
	Files    []string `json:"files" binding:"required"`
	Pattern  string   `json:"pattern" binding:"required"`
	NewValue string   `json:"newValue" binding:"required"`
} // @name ReplaceRequest

type Match struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
} // @name Match
