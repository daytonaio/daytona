// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// RemoteMetrics holds system metrics collected from a remote host
type RemoteMetrics struct {
	CPUUsagePercent    float64
	MemoryUsagePercent float64
	DiskUsagePercent   float64
	TotalCPUs          int
	TotalMemoryGiB     float64
	TotalDiskGiB       float64
}

// GetRemoteMetrics collects system metrics from the remote libvirt host via SSH
// Returns nil if in local mode or if collection fails
func (l *LibVirt) GetRemoteMetrics(ctx context.Context) (*RemoteMetrics, error) {
	if l.isLocalURI() {
		return nil, fmt.Errorf("not in remote mode")
	}

	host := l.extractHostFromURI()
	if host == "" {
		return nil, fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	metrics := &RemoteMetrics{}

	// Collect all metrics in parallel for better performance
	type result struct {
		name  string
		value interface{}
		err   error
	}
	results := make(chan result, 6)

	// CPU usage
	go func() {
		val, err := l.getRemoteCPUUsage(ctx, host)
		results <- result{"cpu_usage", val, err}
	}()

	// Memory usage
	go func() {
		usagePercent, totalGiB, err := l.getRemoteMemoryUsage(ctx, host)
		results <- result{"mem_usage", usagePercent, err}
		results <- result{"mem_total", totalGiB, err}
	}()

	// Disk usage
	go func() {
		usagePercent, totalGiB, err := l.getRemoteDiskUsage(ctx, host)
		results <- result{"disk_usage", usagePercent, err}
		results <- result{"disk_total", totalGiB, err}
	}()

	// CPU count
	go func() {
		val, err := l.getRemoteCPUCount(ctx, host)
		results <- result{"cpu_count", val, err}
	}()

	// Collect results with timeout
	timeout := time.After(30 * time.Second)
	collected := 0
	expected := 6

	for collected < expected {
		select {
		case r := <-results:
			collected++
			if r.err != nil {
				log.Warnf("Failed to collect remote metric %s: %v", r.name, r.err)
				continue
			}
			switch r.name {
			case "cpu_usage":
				metrics.CPUUsagePercent = r.value.(float64)
			case "mem_usage":
				metrics.MemoryUsagePercent = r.value.(float64)
			case "mem_total":
				metrics.TotalMemoryGiB = r.value.(float64)
			case "disk_usage":
				metrics.DiskUsagePercent = r.value.(float64)
			case "disk_total":
				metrics.TotalDiskGiB = r.value.(float64)
			case "cpu_count":
				metrics.TotalCPUs = r.value.(int)
			}
		case <-timeout:
			return metrics, fmt.Errorf("timeout collecting remote metrics")
		case <-ctx.Done():
			return metrics, ctx.Err()
		}
	}

	log.Debugf("Remote metrics collected: CPU=%.2f%%, Mem=%.2f%%, Disk=%.2f%%, CPUs=%d, MemGiB=%.2f, DiskGiB=%.2f",
		metrics.CPUUsagePercent, metrics.MemoryUsagePercent, metrics.DiskUsagePercent,
		metrics.TotalCPUs, metrics.TotalMemoryGiB, metrics.TotalDiskGiB)

	return metrics, nil
}

// getRemoteCPUUsage gets CPU usage percentage from the remote host
func (l *LibVirt) getRemoteCPUUsage(ctx context.Context, host string) (float64, error) {
	// Use vmstat to get CPU idle percentage, then calculate usage
	// vmstat 1 2 gives us a 1-second sample
	cmd := exec.CommandContext(ctx, "ssh", "-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=no", host,
		"vmstat 1 2 | tail -1 | awk '{print $15}'")

	output, err := cmd.Output()
	if err != nil {
		// Fallback: try using /proc/stat
		return l.getRemoteCPUUsageFromProc(ctx, host)
	}

	idleStr := strings.TrimSpace(string(output))
	idle, err := strconv.ParseFloat(idleStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse CPU idle: %w", err)
	}

	// CPU usage = 100 - idle
	return 100.0 - idle, nil
}

// getRemoteCPUUsageFromProc gets CPU usage from /proc/stat (fallback method)
func (l *LibVirt) getRemoteCPUUsageFromProc(ctx context.Context, host string) (float64, error) {
	// Get two samples 1 second apart to calculate CPU usage
	cmd := exec.CommandContext(ctx, "ssh", "-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=no", host,
		`cat /proc/stat | grep '^cpu ' | head -1; sleep 1; cat /proc/stat | grep '^cpu ' | head -1`)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get /proc/stat: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("insufficient /proc/stat data")
	}

	// Parse the two CPU lines
	parse := func(line string) (idle, total int64, err error) {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			return 0, 0, fmt.Errorf("invalid /proc/stat format")
		}
		var vals []int64
		for i := 1; i < len(fields) && i <= 10; i++ {
			v, _ := strconv.ParseInt(fields[i], 10, 64)
			vals = append(vals, v)
			total += v
		}
		if len(vals) >= 4 {
			idle = vals[3] // idle is the 4th field
		}
		return idle, total, nil
	}

	idle1, total1, err := parse(lines[0])
	if err != nil {
		return 0, err
	}
	idle2, total2, err := parse(lines[1])
	if err != nil {
		return 0, err
	}

	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)
	if totalDelta == 0 {
		return 0, nil
	}

	usage := 100.0 * (1.0 - idleDelta/totalDelta)
	return usage, nil
}

// getRemoteMemoryUsage gets memory usage from the remote host
func (l *LibVirt) getRemoteMemoryUsage(ctx context.Context, host string) (usagePercent float64, totalGiB float64, err error) {
	// Use free command to get memory info
	cmd := exec.CommandContext(ctx, "ssh", "-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=no", host,
		"free -b | grep Mem | awk '{print $2, $3, $7}'")

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get memory info: %w", err)
	}

	fields := strings.Fields(strings.TrimSpace(string(output)))
	if len(fields) < 3 {
		return 0, 0, fmt.Errorf("invalid memory info format")
	}

	total, _ := strconv.ParseFloat(fields[0], 64)
	used, _ := strconv.ParseFloat(fields[1], 64)
	available, _ := strconv.ParseFloat(fields[2], 64)

	// Calculate usage based on (total - available) / total
	// This matches how gopsutil calculates it
	if total > 0 {
		usagePercent = ((total - available) / total) * 100.0
		totalGiB = total / (1024 * 1024 * 1024)
	}

	_ = used // Suppress unused variable warning
	return usagePercent, totalGiB, nil
}

// getRemoteDiskUsage gets disk usage from the remote host (root filesystem)
func (l *LibVirt) getRemoteDiskUsage(ctx context.Context, host string) (usagePercent float64, totalGiB float64, err error) {
	// Use df with --output for cleaner parsing, specifically targeting the root mount point
	// The command filters to only show the line for "/" mount point to avoid confusion with other mounts
	cmd := exec.CommandContext(ctx, "ssh", "-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=no", "-T", host,
		`df -B1 --output=target,size,used,pcent 2>/dev/null | awk '$1 == "/" {gsub(/%/, "", $4); print $2, $3, $4}'`)

	output, err := cmd.Output()
	if err != nil {
		// Fallback to simpler command if --output is not supported
		return l.getRemoteDiskUsageFallback(ctx, host)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return l.getRemoteDiskUsageFallback(ctx, host)
	}

	fields := strings.Fields(outputStr)
	if len(fields) < 3 {
		log.Warnf("Unexpected disk info format: %q, trying fallback", outputStr)
		return l.getRemoteDiskUsageFallback(ctx, host)
	}

	total, _ := strconv.ParseFloat(fields[0], 64)
	used, _ := strconv.ParseFloat(fields[1], 64)
	usagePercent, _ = strconv.ParseFloat(fields[2], 64)

	// If percentage parsing failed, calculate it from total/used
	if usagePercent == 0 && total > 0 && used > 0 {
		usagePercent = (used / total) * 100.0
	}

	if total > 0 {
		totalGiB = total / (1024 * 1024 * 1024)
	}

	log.Debugf("Remote disk usage: total=%v, used=%v, percent=%.2f%%", total, used, usagePercent)
	return usagePercent, totalGiB, nil
}

// getRemoteDiskUsageFallback is a fallback method for systems without df --output support
func (l *LibVirt) getRemoteDiskUsageFallback(ctx context.Context, host string) (usagePercent float64, totalGiB float64, err error) {
	// Use stat to get filesystem info - more reliable across different systems
	cmd := exec.CommandContext(ctx, "ssh", "-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=no", "-T", host,
		`stat -f -c '%b %a %S' /`)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get disk info via stat: %w", err)
	}

	fields := strings.Fields(strings.TrimSpace(string(output)))
	if len(fields) < 3 {
		return 0, 0, fmt.Errorf("invalid stat output format")
	}

	totalBlocks, _ := strconv.ParseFloat(fields[0], 64)
	availBlocks, _ := strconv.ParseFloat(fields[1], 64)
	blockSize, _ := strconv.ParseFloat(fields[2], 64)

	totalBytes := totalBlocks * blockSize
	availBytes := availBlocks * blockSize
	usedBytes := totalBytes - availBytes

	if totalBytes > 0 {
		usagePercent = (usedBytes / totalBytes) * 100.0
		totalGiB = totalBytes / (1024 * 1024 * 1024)
	}

	log.Debugf("Remote disk usage (fallback): total=%.0f, avail=%.0f, percent=%.2f%%", totalBytes, availBytes, usagePercent)
	return usagePercent, totalGiB, nil
}

// getRemoteCPUCount gets the number of CPUs from the remote host
func (l *LibVirt) getRemoteCPUCount(ctx context.Context, host string) (int, error) {
	cmd := exec.CommandContext(ctx, "ssh", "-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=no", host,
		"nproc")

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get CPU count: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse CPU count: %w", err)
	}

	return count, nil
}

// IsRemoteMode returns true if the libvirt connection is to a remote host
func (l *LibVirt) IsRemoteMode() bool {
	return !l.isLocalURI()
}
