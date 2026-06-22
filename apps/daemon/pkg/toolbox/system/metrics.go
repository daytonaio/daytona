// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package system

import (
	"context"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/daytonaio/common-go/pkg/telemetry"
	"github.com/gin-gonic/gin"
)

const sampleInterval = 5 * time.Second

type sampler struct {
	limits *telemetry.ResourceLimits

	mu          sync.RWMutex
	cpuUsedPct  float64
	lastCPU     uint64
	lastSampled time.Time
}

func NewSampler() *sampler {
	return &sampler{
		limits: telemetry.GetContainerLimits(),
	}
}

// Start runs the sampling loop until ctx is cancelled. Each tick computes CPU
// utilization as a delta against the previous reading, so cpuUsedPct reflects
// the average over the last sample window. Memory and disk are point reads
// served on demand by the handler, so they are not sampled here.
func (s *sampler) Start(ctx context.Context) {
	ticker := time.NewTicker(sampleInterval)
	defer ticker.Stop()

	// Establish a baseline so the first tick can produce a delta; until then
	// cpuUsedPct stays 0 (cold start).
	s.sample()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sample()
		}
	}
}

func (s *sampler) sample() {
	cpuUsage, err := telemetry.ReadCgroupCPUUsageNanos(s.limits.CgroupV2())
	if err != nil {
		slog.Debug("system metrics: failed to read cgroup CPU usage", "error", err)
		return
	}
	s.update(cpuUsage, time.Now(), s.limits.CPULimit)
}

// update folds a fresh cumulative CPU reading into the cached cpuUsedPct. It is
// split from sample() and takes cpuLimit explicitly so the delta state machine
// can be unit-tested without reading cgroup files.
func (s *sampler) update(cpuUsage uint64, now time.Time, cpuLimit float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.lastSampled.IsZero() && cpuLimit > 0 {
		wallDelta := now.Sub(s.lastSampled).Nanoseconds()
		if cpuUsage >= s.lastCPU && wallDelta > 0 {
			s.cpuUsedPct = telemetry.CPUUsagePercent(cpuUsage-s.lastCPU, wallDelta, cpuLimit)
		} else {
			// Counter reset (cgroup recreated) or non-monotonic clock: the prior
			// delta is meaningless, so stop reporting the stale value.
			s.cpuUsedPct = 0
		}
	}
	s.lastCPU = cpuUsage
	s.lastSampled = now
}

// GetSystemMetrics godoc
//
//	@Summary		Get sandbox resource metrics
//	@Description	Current CPU/memory/disk usage snapshot for the sandbox. cpuUsedPct is the
//	@Description	average CPU usage as a percentage of the CPU limit over the last sample
//	@Description	window (0 until the first sample completes). Byte fields are in bytes.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	SystemMetrics
//	@Router			/system/metrics [get]
//
//	@id				GetSystemMetrics
func (s *sampler) GetSystemMetrics(c *gin.Context) {
	s.mu.RLock()
	cpuUsedPct := s.cpuUsedPct
	s.mu.RUnlock()

	cpuCount := int(math.Ceil(s.limits.CPULimit))

	now := time.Now().UTC()
	metrics := SystemMetrics{
		Timestamp:     now.Format(time.RFC3339),
		TimestampUnix: now.Unix(),
		CpuCount:      cpuCount,
		CpuUsedPct:    cpuUsedPct,
		MemTotal:      int64(s.limits.MemoryLimit),
	}

	if memUsed, err := telemetry.ReadCgroupMemUsageBytes(s.limits.CgroupV2()); err != nil {
		slog.Debug("system metrics: failed to read memory usage", "error", err)
	} else {
		metrics.MemUsed = int64(memUsed)
	}

	if memCache, err := telemetry.ReadCgroupMemCacheBytes(s.limits.CgroupV2()); err != nil {
		slog.Debug("system metrics: failed to read memory cache", "error", err)
	} else {
		metrics.MemCache = int64(memCache)
	}

	if disk, err := telemetry.GetDiskStats("/"); err != nil {
		slog.Debug("system metrics: failed to read disk stats", "error", err)
	} else {
		metrics.DiskUsed = int64(disk.Used)
		metrics.DiskTotal = int64(disk.Total)
		metrics.DiskFree = int64(disk.Available)
	}

	c.JSON(http.StatusOK, metrics)
}
