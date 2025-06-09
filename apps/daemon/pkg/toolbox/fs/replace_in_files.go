// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func ReplaceInFiles(c *gin.Context) {
	var req ReplaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	results := make([]ReplaceResult, 0, len(req.Files))

	for _, filePath := range req.Files {
		content, err := os.ReadFile(filePath)
		if err != nil {
			results = append(results, ReplaceResult{
				File:    filePath,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		newValue := ""
		if req.NewValue != nil {
			newValue = *req.NewValue
		}

		newContent := strings.ReplaceAll(string(content), req.Pattern, newValue)

		err = os.WriteFile(filePath, []byte(newContent), 0644)
		if err != nil {
			results = append(results, ReplaceResult{
				File:    filePath,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		results = append(results, ReplaceResult{
			File:    filePath,
			Success: true,
		})
	}

	c.JSON(http.StatusOK, results)
}
