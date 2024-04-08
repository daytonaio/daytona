// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// GetGitUser 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git context
//	@Description	Get Git context
//	@Produce		json
//	@Param			gitProviderId	path		string	true	"Git Provider Id"
//	@Success		200				{object}	GitUser
//	@Router			/gitprovider/{gitProviderId}/user [get]
//
//	@id				GetGitUser
func GetGitUser(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")

	server := server.GetInstance(nil)

	response, err := server.GitProviderService.GetGitUser(gitProviderId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get git user: %s", err.Error()))
		return
	}

	ctx.JSON(200, response)
}
