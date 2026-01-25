// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// MemoryStatsResponse represents memory statistics from /proc/meminfo
type MemoryStatsResponse struct {
	MemTotalKiB     uint64 `json:"memTotalKiB"`
	MemFreeKiB      uint64 `json:"memFreeKiB"`
	MemAvailableKiB uint64 `json:"memAvailableKiB"`
	BuffersKiB      uint64 `json:"buffersKiB"`
	CachedKiB       uint64 `json:"cachedKiB"`
} // @name MemoryStatsResponse

// GetMemoryStats godoc
//
//	@Summary		Get memory statistics
//	@Description	Get current memory usage statistics from /proc/meminfo
//	@Tags			info
//	@Produce		json
//	@Success		200	{object}	MemoryStatsResponse
//	@Failure		500	{object}	map[string]string
//	@Router			/memory-stats [get]
//
//	@id				GetMemoryStats
func (s *Server) GetMemoryStats(ctx *gin.Context) {
	stats, err := parseMemInfo()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, stats)
}

// parseMemInfo reads and parses /proc/meminfo
func parseMemInfo() (*MemoryStatsResponse, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/meminfo: %w", err)
	}
	defer file.Close()

	stats := &MemoryStatsResponse{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// Remove the trailing colon from the key
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch key {
		case "MemTotal":
			stats.MemTotalKiB = value
		case "MemFree":
			stats.MemFreeKiB = value
		case "MemAvailable":
			stats.MemAvailableKiB = value
		case "Buffers":
			stats.BuffersKiB = value
		case "Cached":
			stats.CachedKiB = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/meminfo: %w", err)
	}

	return stats, nil
}
