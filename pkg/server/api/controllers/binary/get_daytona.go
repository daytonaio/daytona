// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package binary

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/internal/constants"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/frpc"
	"github.com/gin-gonic/gin"
)

// Used in projects to download the Daytona binary
func GetDaytonaScript(ctx *gin.Context) {
	c, err := config.GetConfig()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get config: %s", err.Error()))
		return
	}

	downloadUrl, _ := url.JoinPath(frpc.GetApiUrl(c), "binary")
	getServerScript := constants.GetDaytonaScript(downloadUrl)

	ctx.String(http.StatusOK, getServerScript)
}
