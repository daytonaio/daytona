// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"time"
)

type GitCommitRequest struct {
	Path    string `json:"path" binding:"required"`
	Message string `json:"message" binding:"required"`
	Author  string `json:"author" binding:"required"`
	Email   string `json:"email" binding:"required"`
} // @name GitCommitRequest

type GitPushRequest struct {
	Path     string `json:"path" binding:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
} // @name GitPushRequest

type GitBranchRequest struct {
	Path string `json:"path" binding:"required"`
	Name string `json:"name" binding:"required"`
} // @name GitBranchRequest

type GitCommitInfo struct {
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
} // @name GitCommitInfo
