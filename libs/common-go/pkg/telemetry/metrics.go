// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/host"
	otel_runtime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdk_metric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	log "github.com/sirupsen/logrus"
)

// InitMetrics initializes OpenTelemetry Metrics with an OTLP HTTP exporter.
func InitMetrics(ctx context.Context, config Config, meterName string) (*sdk_metric.MeterProvider, error) {
	// Resource describing this service
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create OTLP HTTP metrics exporter
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(config.Endpoint+"/v1/metrics"),
		otlpmetrichttp.WithHeaders(config.Headers),
	)
	if err != nil {
		return nil, err
	}

	// Periodic reader to push metrics on an interval
	reader := sdk_metric.NewPeriodicReader(exporter)

	// MeterProvider with resource and reader
	mp := sdk_metric.NewMeterProvider(
		sdk_metric.WithResource(res),
		sdk_metric.WithReader(reader),
	)

	// Set as global provider so otel.Meter(...) uses it
	otel.SetMeterProvider(mp)

	// Get container limits
	limits := getContainerLimits()
	log.Infof("Detected container limits - CPU: %.2f cores, Memory: %d bytes (%.2f GB)",
		limits.CPULimit,
		limits.MemoryLimit,
		float64(limits.MemoryLimit)/1073741824.0)

	// ADDED: Log initial disk stats for verification
	if diskStats, err := getDiskStats("/"); err == nil {
		log.Infof("Detected filesystem - Total: %.2f GB, Used: %.2f GB, Available: %.2f GB",
			float64(diskStats.Total)/1073741824.0,
			float64(diskStats.Used)/1073741824.0,
			float64(diskStats.Available)/1073741824.0)
	}

	// Register container limits metrics
	if err := registerLimitsMetrics(meterName, mp, limits); err != nil {
		log.Warnf("Failed to register container limits metrics: %v", err)
	}

	// Register container usage metrics
	if err := registerUsageMetrics(meterName, mp, limits); err != nil {
		log.Warnf("Failed to register container usage metrics: %v", err)
	}

	// Register disk usage metrics
	if err := registerDiskUsageMetrics(meterName, mp); err != nil {
		log.Warnf("Failed to register disk usage metrics: %v", err)
	}

	// Start runtime metrics collection
	if err := otel_runtime.Start(otel_runtime.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
		log.Warnf("Failed to start runtime metrics: %v", err)
	}

	// Start host metrics collection
	log.Info("Starting host metrics collection")
	if err := host.Start(host.WithMeterProvider(mp)); err != nil {
		log.Warnf("Failed to start host metrics: %v", err)
	}

	return mp, nil
}

// ShutdownMeter gracefully shuts down the MeterProvider and flushes metrics.
func ShutdownMeter(mp *sdk_metric.MeterProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mp.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down meter provider: %v", err)
	}
}

type ResourceLimits struct {
	CPULimit    float64 // CPU cores
	MemoryLimit uint64  // bytes
	cgroupV2    bool
}

type DiskStats struct {
	Total     uint64 // Total bytes
	Used      uint64 // Used bytes
	Available uint64 // Available bytes
	Path      string // Mount path
}

// detectCgroupVersion detects if system is using cgroup v1 or v2
func detectCgroupVersion() bool {
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err == nil {
		return true // cgroup v2
	}
	return false // cgroup v1
}

// getContainerLimits reads container resource limits from cgroups
func getContainerLimits() *ResourceLimits {
	limits := &ResourceLimits{
		cgroupV2: detectCgroupVersion(),
	}

	if limits.cgroupV2 {
		limits.readCgroupV2Limits()
	} else {
		limits.readCgroupV1Limits()
	}

	return limits
}

func (cl *ResourceLimits) readCgroupV1Limits() {
	// Read CPU limit from cgroup v1
	quota, _ := readInt64FromFile("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
	period, _ := readInt64FromFile("/sys/fs/cgroup/cpu/cpu.cfs_period_us")

	if quota > 0 && period > 0 {
		cl.CPULimit = float64(quota) / float64(period)
	} else {
		cl.CPULimit = float64(runtime.NumCPU())
	}

	// Read memory limit from cgroup v1
	if memLimit, err := readUint64FromFile("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil {
		// Check if limit is set to a reasonable value (not max uint64)
		if memLimit < 0x7FFFFFFFFFFFF000 {
			cl.MemoryLimit = memLimit
		}
	}
}

func (cl *ResourceLimits) readCgroupV2Limits() {
	// Read CPU limit from cgroup v2
	if data, err := os.ReadFile("/sys/fs/cgroup/cpu.max"); err == nil {
		parts := strings.Fields(strings.TrimSpace(string(data)))
		if len(parts) >= 2 && parts[0] != "max" {
			quota, _ := strconv.ParseInt(parts[0], 10, 64)
			period, _ := strconv.ParseInt(parts[1], 10, 64)
			if quota > 0 && period > 0 {
				cl.CPULimit = float64(quota) / float64(period)
			}
		}
	}

	if cl.CPULimit == 0 {
		cl.CPULimit = float64(runtime.NumCPU())
	}

	// Read memory limit from cgroup v2
	if memLimit, err := readUint64FromFile("/sys/fs/cgroup/memory.max"); err == nil {
		if memLimit < 0x7FFFFFFFFFFFF000 {
			cl.MemoryLimit = memLimit
		}
	}
}

func readInt64FromFile(filepath string) (int64, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}

func readUint64FromFile(filepath string) (uint64, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}

func getDiskStats(path string) (*DiskStats, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, err
	}

	total := stat.Blocks * uint64(stat.Bsize)
	available := stat.Bavail * uint64(stat.Bsize)
	used := total - (stat.Bfree * uint64(stat.Bsize))

	return &DiskStats{
		Total:     total,
		Used:      used,
		Available: available,
		Path:      path,
	}, nil
}

// registerLimitsMetrics registers gauges for resource limits
func registerLimitsMetrics(name string, mp *sdk_metric.MeterProvider, limits *ResourceLimits) error {
	meter := mp.Meter(fmt.Sprintf("%s-limits", name))

	// CPU Limit gauge
	cpuLimitGauge, err := meter.Float64ObservableGauge(
		fmt.Sprintf("%s.cpu.limit", name),
		metric.WithDescription("CPU limit in cores"),
		metric.WithUnit("cores"),
	)
	if err != nil {
		return err
	}

	// Memory Limit gauge
	memoryLimitGauge, err := meter.Int64ObservableGauge(
		fmt.Sprintf("%s.memory.limit", name),
		metric.WithDescription("Memory limit in bytes"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return err
	}

	// Register callback to observe the limits
	_, err = meter.RegisterCallback(
		func(ctx context.Context, observer metric.Observer) error {
			if limits.CPULimit > 0 {
				observer.ObserveFloat64(cpuLimitGauge, limits.CPULimit)
			}
			if limits.MemoryLimit > 0 {
				observer.ObserveInt64(memoryLimitGauge, int64(limits.MemoryLimit))
			}
			return nil
		},
		cpuLimitGauge,
		memoryLimitGauge,
	)

	return err
}

// registerUsageMetrics registers gauges for resource usage
func registerUsageMetrics(name string, mp *sdk_metric.MeterProvider, limits *ResourceLimits) error {
	meter := mp.Meter(fmt.Sprintf("%s-usage", name))

	// CPU Usage Percentage (relative to container limit)
	cpuUtilGauge, err := meter.Float64ObservableGauge(
		fmt.Sprintf("%s.cpu.utilization", name),
		metric.WithDescription("CPU utilization as percentage of limit"),
		metric.WithUnit("percent"),
	)
	if err != nil {
		return err
	}

	// Memory Usage Percentage (relative to container limit)
	memUtilGauge, err := meter.Float64ObservableGauge(
		fmt.Sprintf("%s.memory.utilization", name),
		metric.WithDescription("Memory utilization as percentage of limit"),
		metric.WithUnit("percent"),
	)
	if err != nil {
		return err
	}

	// Memory Usage (absolute)
	memUsageGauge, err := meter.Int64ObservableGauge(
		fmt.Sprintf("%s.memory.usage", name),
		metric.WithDescription("Memory usage in bytes"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return err
	}

	var lastCPUUsage uint64
	var lastTimestamp time.Time

	_, err = meter.RegisterCallback(
		func(ctx context.Context, observer metric.Observer) error {
			// Read current memory usage
			var memUsage uint64
			if limits.cgroupV2 {
				memUsage, _ = readUint64FromFile("/sys/fs/cgroup/memory.current")
			} else {
				memUsage, _ = readUint64FromFile("/sys/fs/cgroup/memory/memory.usage_in_bytes")
			}

			if memUsage > 0 {
				observer.ObserveInt64(memUsageGauge, int64(memUsage))

				// Calculate memory utilization percentage
				if limits.MemoryLimit > 0 {
					memUtilPercent := float64(memUsage) / float64(limits.MemoryLimit) * 100.0
					observer.ObserveFloat64(memUtilGauge, memUtilPercent)
				}
			}

			// Read current CPU usage
			var cpuUsage uint64
			if limits.cgroupV2 {
				// For cgroup v2, read usage_usec from cpu.stat
				if data, err := os.ReadFile("/sys/fs/cgroup/cpu.stat"); err == nil {
					lines := strings.Split(string(data), "\n")
					for _, line := range lines {
						if strings.HasPrefix(line, "usage_usec") {
							parts := strings.Fields(line)
							if len(parts) >= 2 {
								usec, _ := strconv.ParseUint(parts[1], 10, 64)
								cpuUsage = usec * 1000 // Convert to nanoseconds
								break
							}
						}
					}
				}
			} else {
				cpuUsage, _ = readUint64FromFile("/sys/fs/cgroup/cpu/cpuacct.usage")
			}

			// Calculate CPU utilization
			now := time.Now()
			if !lastTimestamp.IsZero() && cpuUsage > lastCPUUsage && limits.CPULimit > 0 {
				timeDelta := now.Sub(lastTimestamp).Nanoseconds()
				cpuDelta := cpuUsage - lastCPUUsage

				// CPU usage as percentage of allocated CPUs
				cpuUsagePercent := (float64(cpuDelta) / float64(timeDelta)) / limits.CPULimit * 100.0
				observer.ObserveFloat64(cpuUtilGauge, cpuUsagePercent)
			}

			lastCPUUsage = cpuUsage
			lastTimestamp = now

			return nil
		},
		cpuUtilGauge,
		memUtilGauge,
		memUsageGauge,
	)

	return err
}

// registerDiskUsageMetrics registers gauges for filesystem/disk usage metrics
func registerDiskUsageMetrics(name string, mp *sdk_metric.MeterProvider) error {
	meter := mp.Meter(fmt.Sprintf("%s-disk", name))

	// Disk usage in bytes
	diskUsageGauge, err := meter.Int64ObservableGauge(
		fmt.Sprintf("%s.filesystem.usage", name),
		metric.WithDescription("Filesystem usage in bytes"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return err
	}

	// Disk total size in bytes
	diskTotalGauge, err := meter.Int64ObservableGauge(
		fmt.Sprintf("%s.filesystem.total", name),
		metric.WithDescription("Total filesystem size in bytes"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return err
	}

	// Disk available in bytes
	diskAvailableGauge, err := meter.Int64ObservableGauge(
		fmt.Sprintf("%s.filesystem.available", name),
		metric.WithDescription("Available filesystem space in bytes"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return err
	}

	// Disk utilization percentage
	diskUtilGauge, err := meter.Float64ObservableGauge(
		fmt.Sprintf("%s.filesystem.utilization", name),
		metric.WithDescription("Filesystem utilization as percentage"),
		metric.WithUnit("percent"),
	)
	if err != nil {
		return err
	}

	// Register callback to collect disk metrics
	_, err = meter.RegisterCallback(
		func(ctx context.Context, observer metric.Observer) error {
			// Get disk stats for root filesystem
			diskStats, err := getDiskStats("/")
			if err != nil {
				log.Warnf("Failed to get disk stats: %v", err)
				return nil // Don't fail the entire callback
			}

			// Create attributes for the filesystem
			attrs := metric.WithAttributes(
				attribute.String("device", diskStats.Path),
			)

			// Observe metrics
			observer.ObserveInt64(diskUsageGauge, int64(diskStats.Used), attrs)
			observer.ObserveInt64(diskTotalGauge, int64(diskStats.Total), attrs)
			observer.ObserveInt64(diskAvailableGauge, int64(diskStats.Available), attrs)

			// Calculate utilization percentage
			if diskStats.Total > 0 {
				utilizationPercent := float64(diskStats.Used) / float64(diskStats.Total) * 100.0
				observer.ObserveFloat64(diskUtilGauge, utilizationPercent, attrs)
			}

			return nil
		},
		diskUsageGauge,
		diskTotalGauge,
		diskAvailableGauge,
		diskUtilGauge,
	)

	return err
}
