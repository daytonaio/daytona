// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package binary

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// Serves the Daytona binary based on the requested version and name.
// The name of the binary follows the pattern: daytona-<os>-<arch>[.exe]
func GetBinary(ctx *gin.Context) {
	binaryVersion := ctx.Param("version")
	binaryName := ctx.Param("binaryName")

	server := server.GetInstance(nil)
	binaryPath, err := server.GetBinaryPath(binaryName, binaryVersion)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get binary path: %s", err.Error()))
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(binaryPath)))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(binaryPath)
}
