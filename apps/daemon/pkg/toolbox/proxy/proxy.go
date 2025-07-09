// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
)

func GetProxyTarget(ctx *gin.Context) (*url.URL, map[string]string, error) {
	targetPort := ctx.Param("port")
	if targetPort == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("target port is required")))
		return nil, nil, errors.New("target port is required")
	}

	// Build the target URL
	targetURL := fmt.Sprintf("http://localhost:%s", targetPort)

	// Get the wildcard path and normalize it
	path := ctx.Param("path")

	// Ensure path always has a leading slash but not duplicate slashes
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	target, err := url.Parse(fmt.Sprintf("%s%s", targetURL, path))
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err)))
		return nil, nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	return target, nil, nil
}
