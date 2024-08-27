// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package binary

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/internal/constants"
	"github.com/gin-gonic/gin"
)

// Used in projects to download the Daytona binary
func GetDaytonaScript(ctx *gin.Context) {
	scheme := "http"
	if ctx.Request.TLS != nil || ctx.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	downloadUrl, _ := url.JoinPath(fmt.Sprintf("%s://%s", scheme, ctx.Request.Host), "binary")
	getServerScript := constants.GetDaytonaScript(downloadUrl)

	ctx.String(http.StatusOK, getServerScript)
}
