//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import "github.com/gin-gonic/gin"

func (s *server) registerPlatformRoutes(r *gin.Engine) {}

func (s *server) shutdownPlatform() {}
