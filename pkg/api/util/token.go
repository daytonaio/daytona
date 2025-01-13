// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func ExtractToken(ctx *gin.Context) string {
	bearerToken := ctx.GetHeader("Authorization")
	if bearerToken == "" {
		return ""
	}

	if !strings.HasPrefix(bearerToken, "Bearer ") {
		return ""
	}

	return strings.TrimPrefix(bearerToken, "Bearer ")
}
