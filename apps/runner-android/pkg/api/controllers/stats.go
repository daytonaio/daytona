// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// MemoryStatsResponse represents the API response for memory stats
type MemoryStatsResponse struct {
	SandboxIds []string  `json:"sandbox_ids"`
	Stats      []any     `json:"stats"`
	FromTime   time.Time `json:"from_time"`
	ToTime     time.Time `json:"to_time"`
	Count      int       `json:"count"`
}

// GetMemoryStatsJSON returns memory statistics as JSON
// Note: Memory ballooning is not supported for Cuttlefish
//
//	@Summary		Get memory statistics
//	@Description	Returns memory statistics for sandboxes over a time range (not supported for Cuttlefish)
//	@Tags			stats
//	@Produce		json
//	@Param			sandbox	query		string	false	"Filter by sandbox ID"
//	@Param			hours	query		int		false	"Lookback hours (default: 24)"
//	@Success		200		{object}	MemoryStatsResponse
//	@Failure		500		{object}	string
//	@Router			/stats/memory [get]
//
//	@id				GetMemoryStatsJSON
func GetMemoryStatsJSON(ctx *gin.Context) {
	// Memory statistics are not supported for Cuttlefish
	// Android instances don't have the same memory ballooning concept
	response := MemoryStatsResponse{
		SandboxIds: []string{},
		Stats:      []any{},
		FromTime:   time.Now(),
		ToTime:     time.Now(),
		Count:      0,
	}

	ctx.JSON(http.StatusOK, response)
}

// GetMemoryStatsViewHTML returns an HTML visualization page for memory stats
// Note: Memory ballooning is not supported for Cuttlefish
//
//	@Summary		Memory stats visualization
//	@Description	Returns an interactive HTML page with memory usage charts (not supported for Cuttlefish)
//	@Tags			stats
//	@Produce		html
//	@Param			hours	query		int		false	"Lookback hours (default: 24)"
//	@Success		200		{string}	string	"HTML page"
//	@Router			/stats/memory/view [get]
//
//	@id				GetMemoryStatsViewHTML
func GetMemoryStatsViewHTML(ctx *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Cuttlefish Memory Stats</title>
	<style>
		body {
			font-family: system-ui, -apple-system, sans-serif;
			max-width: 800px;
			margin: 50px auto;
			padding: 20px;
			text-align: center;
		}
		.message {
			background: #f0f0f0;
			padding: 20px;
			border-radius: 8px;
			margin-top: 20px;
		}
	</style>
</head>
<body>
	<h1>Memory Statistics</h1>
	<div class="message">
		<p>Memory ballooning and statistics are not supported for Cuttlefish Android instances.</p>
		<p>Android virtual devices manage their own memory allocation.</p>
	</div>
</body>
</html>`
	ctx.Header("Content-Type", "text/html")
	ctx.String(http.StatusOK, html)
}
