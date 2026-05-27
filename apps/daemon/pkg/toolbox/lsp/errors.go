// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package lsp

import (
	"errors"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
)

var ErrLspServerNotInitialized = errors.New("server not initialized")

func classifyLspError(err error) error {
	if errors.Is(err, ErrLspServerNotInitialized) {
		return common.NewLspServerNotInitializedError(err.Error())
	}
	return common_errors.NewBadRequestError(err)
}

func abortWithLspError(c *gin.Context, err error) {
	c.Error(classifyLspError(err))
}
