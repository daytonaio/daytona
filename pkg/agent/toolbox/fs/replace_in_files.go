// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func ReplaceInFiles(c *gin.Context) {
	var req ReplaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	type Result struct {
		File    string `json:"file"`
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}

	results := make([]Result, 0, len(req.Files))

	for _, filePath := range req.Files {
		content, err := os.ReadFile(filePath)
		if err != nil {
			results = append(results, Result{
				File:    filePath,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		newContent := strings.ReplaceAll(string(content), req.Pattern, req.NewValue)

		err = os.WriteFile(filePath, []byte(newContent), 0644)
		if err != nil {
			results = append(results, Result{
				File:    filePath,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		results = append(results, Result{
			File:    filePath,
			Success: true,
		})
	}

	c.JSON(200, results)
}
