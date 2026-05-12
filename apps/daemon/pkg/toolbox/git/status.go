// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"net/http"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// ─── Approach B: per-endpoint per-status typed error schemas ──────────────────
// Each schema has a narrow code enum — the model type alone identifies the error.

// GitStatusErrorCode lists codes returned by the GET /git/status endpoint.
type GitStatusErrorCode string //	@name	GitStatusErrorCode

const (
	GitStatusCodeRepoNotFound    GitStatusErrorCode = "GIT_REPO_NOT_FOUND"
	GitStatusCodeInternalError   GitStatusErrorCode = "INTERNAL_SERVER_ERROR"
)

// GitStatusNotFoundError is returned as 404 by GET /git/status.
//
//	@Description	Returned when no git repository exists at the given path
type GitStatusNotFoundError struct {
	StatusCode int                `json:"statusCode" example:"404"`
	Source     string             `json:"source"     example:"DAYTONA_DAEMON"`
	Code       GitStatusErrorCode `json:"code"       example:"GIT_REPO_NOT_FOUND"`
	Message    string             `json:"message"    example:"repository does not exist"`
	Timestamp  time.Time          `json:"timestamp"`
	Path       string             `json:"path"`
	Method     string             `json:"method"`
} //	@name	GitStatusNotFoundError

// GitStatusInternalError is returned as 500 by GET /git/status.
//
//	@Description	Returned for unexpected errors in GET /git/status
type GitStatusInternalError struct {
	StatusCode int                `json:"statusCode" example:"500"`
	Source     string             `json:"source"     example:"DAYTONA_DAEMON"`
	Code       GitStatusErrorCode `json:"code"       example:"INTERNAL_SERVER_ERROR"`
	Message    string             `json:"message"    example:"internal server error"`
	Timestamp  time.Time          `json:"timestamp"`
	Path       string             `json:"path"`
	Method     string             `json:"method"`
} //	@name	GitStatusInternalError

// ─── Handler ──────────────────────────────────────────────────────────────────

// GetStatus godoc
//
//	@Summary		Get Git status
//	@Description	Get the Git status of the repository at the specified path
//	@Tags			git
//	@Produce		json
//	@Param			path	query		string	true	"Repository path"
//	@Success		200		{object}	git.GitStatus
//	@Failure		404		{object}	GitStatusNotFoundError
//	@Failure		500		{object}	GitStatusInternalError
//	@Router			/git/status [get]
//
//	@id				GetStatus
func GetStatus(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		_ = c.Error(common_errors.NewBadRequestError(errors.New("path is required")))
		return
	}

	gitService := git.Service{
		WorkDir: path,
	}

	status, err := gitService.GetGitStatus()
	if err != nil {
		abortWithGitError(c, err)
		return
	}

	c.JSON(http.StatusOK, status)
}
