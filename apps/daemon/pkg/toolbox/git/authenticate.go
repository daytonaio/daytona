// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// Authenticate godoc
//
//	@Summary		Authenticate Git
//	@Description	Persist Git credentials globally via the credential store. Stores the password in plaintext on disk.
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitAuthenticateRequest	true	"Authenticate request"
//	@Success		200
//	@Router			/git/credentials [post]
//
//	@id				Authenticate
func Authenticate(c *gin.Context) {
	var req GitAuthenticateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common_errors.NewInvalidBodyRequestError(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	gitService := git.Service{}

	host := ""
	if req.Host != nil {
		host = *req.Host
	}
	protocol := ""
	if req.Protocol != nil {
		protocol = *req.Protocol
	}

	// The git credential record is line-based ("key=value\n"); a newline or NUL in
	// any field could inject extra attributes or terminate the record early.
	for _, field := range []string{req.Username, req.Password, host, protocol} {
		if strings.ContainsAny(field, "\n\r\x00") {
			_ = c.Error(common_errors.NewBadRequestError(
				errors.New("credential fields must not contain newline or NUL characters")))
			return
		}
	}

	if err := gitService.Authenticate(req.Username, req.Password, host, protocol); err != nil {
		abortWithGitError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
