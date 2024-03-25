// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/gin-gonic/gin"
)

// GetGitUserData 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git context
//	@Description	Get Git context
//	@Produce		json
//	@Param			gitProviderId	path		string	true	"Git Provider Id"
//	@Success		200				{object}	GitUserData
//	@Router			/gitprovider/{gitProviderId}/user-data [get]
//
//	@id				GetGitUserData
func GetGitUserData(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")

	c, err := config.GetConfig()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get config: %s", err.Error()))
		return
	}

	gitProvider := gitprovider.GetGitProvider(gitProviderId, c.GitProviders)
	if gitProvider == nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("git provider not found"))
		return
	}

	userData, err := gitProvider.GetUserData()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get user data: %s", err.Error()))
		return
	}

	ctx.JSON(200, userData)
}
