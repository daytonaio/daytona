// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/controllers"
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
		statusCode, message, codeErr := controllers.GetHTTPStatusCodeAndMessageFromError(err)
		if codeErr != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("%s", message))
		return
	}

	ctx.JSON(200, response)
}
