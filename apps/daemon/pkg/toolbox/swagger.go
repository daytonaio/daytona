// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build !noswagger

package toolbox

import (
	"os"

	"github.com/daytonaio/daemon/internal"
	"github.com/daytonaio/daemon/pkg/toolbox/docs"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func registerSwagger(r *gin.Engine) {
	docs.SwaggerInfo.Description = "Daytona Toolbox API"
	docs.SwaggerInfo.Title = "Daytona Toolbox API"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Version = internal.Version

	if os.Getenv("ENVIRONMENT") != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
}
