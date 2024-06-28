// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package binary

import (
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/internal/constants"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// Used in projects to download the Daytona binary
func GetDaytonaScript(ctx *gin.Context) {
	c, err := server.GetConfig()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	downloadUrl, _ := url.JoinPath(c.GetApiUrl(), "binary")
	getServerScript := constants.GetDaytonaScript(downloadUrl)

	ctx.String(http.StatusOK, getServerScript)
}
