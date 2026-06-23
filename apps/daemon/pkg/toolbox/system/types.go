// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package system

// SystemMetrics is a point-in-time snapshot of a sandbox's CPU, memory and disk
// usage. Byte fields are in bytes; CpuUsedPct is a percentage of the CPU limit.
type SystemMetrics struct {
	Timestamp     string  `json:"timestamp"`
	TimestampUnix int64   `json:"timestampUnix" format:"int64"`
	CpuCount      int     `json:"cpuCount"`
	CpuUsedPct    float64 `json:"cpuUsedPct" format:"double"`
	MemUsed       int64   `json:"memUsed" format:"int64"`
	MemTotal      int64   `json:"memTotal" format:"int64"`
	MemCache      int64   `json:"memCache" format:"int64"`
	DiskUsed      int64   `json:"diskUsed" format:"int64"`
	DiskTotal     int64   `json:"diskTotal" format:"int64"`
	DiskFree      int64   `json:"diskFree" format:"int64"`
} // @name SystemMetrics
