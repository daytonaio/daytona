package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/docker/docker/api/types/image"
)

const systemMetricsKey = "__system_metrics__"

type MetricsServiceConfig struct {
	Endpoint string
	Cache    cache.ICache[models.SystemMetrics]
	Docker   *docker.DockerClient
	Interval time.Duration
}

type MetricsService struct {
	endpoint   string
	httpClient *http.Client
	cache      cache.ICache[models.SystemMetrics]
	docker     *docker.DockerClient
	interval   time.Duration
	cpuHistory *CPUMetricsHistory
}

// CPUMetricsHistory stores previous CPU metrics for delta calculation
type CPUMetricsHistory struct {
	mu          sync.RWMutex
	prevMetrics map[string]float64
	lastUpdate  time.Time
}

// NewPrometheusParser creates a new parser instance
func NewMetricsService(config MetricsServiceConfig) *MetricsService {
	return &MetricsService{
		endpoint: config.Endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache:    config.Cache,
		docker:   config.Docker,
		interval: config.Interval,
		cpuHistory: &CPUMetricsHistory{
			prevMetrics: make(map[string]float64),
		},
	}
}

// FetchMetrics fetches metrics from the Prometheus endpoint
func (s *MetricsService) fetchMetrics() (string, error) {
	resp, err := s.httpClient.Get(s.endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// ParseMetrics parses the Prometheus metrics text format
func (s *MetricsService) parseMetrics(metricsText string) (*models.SystemMetrics, error) {
	metrics := &models.SystemMetrics{}

	// Store raw metric values
	metricValues := make(map[string]float64)
	metricLabels := make(map[string]map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(metricsText))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse metric line - find the last space-separated value as the metric value
		lastSpaceIndex := strings.LastIndex(line, " ")
		if lastSpaceIndex == -1 {
			continue
		}

		metricLabel := strings.TrimSpace(line[:lastSpaceIndex])
		metricValue := strings.TrimSpace(line[lastSpaceIndex+1:])

		// Parse the value
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			continue
		}

		// Parse labels if present
		labels := make(map[string]string)
		metricName := metricLabel

		if strings.Contains(metricLabel, "{") {
			labelStart := strings.Index(metricLabel, "{")
			labelEnd := strings.LastIndex(metricLabel, "}")
			if labelEnd > labelStart {
				labelStr := metricLabel[labelStart+1 : labelEnd]
				actualMetricName := metricLabel[:labelStart]

				// Parse labels - handle comma separation more carefully
				labelPairs := s.splitLabels(labelStr)
				for _, pair := range labelPairs {
					pair = strings.TrimSpace(pair)
					equalIndex := strings.Index(pair, "=")
					if equalIndex > 0 {
						key := strings.TrimSpace(pair[:equalIndex])
						value := strings.TrimSpace(pair[equalIndex+1:])
						// Remove quotes if present
						key = strings.Trim(key, `"`)
						value = strings.Trim(value, `"`)
						labels[key] = value
					}
				}
				metricName = actualMetricName
			}
		}

		// Store metric value and labels
		fullKey := metricName
		if len(labels) > 0 {
			// Create a unique key with labels for complex metrics
			labelParts := make([]string, 0, len(labels))
			for k, v := range labels {
				labelParts = append(labelParts, k+"="+v)
			}
			fullKey = metricName + "{" + strings.Join(labelParts, ",") + "}"
		}

		metricValues[fullKey] = value
		metricLabels[fullKey] = labels
	}

	// Parse and calculate all metrics
	if err := s.parseAllMetrics(metrics, metricValues, metricLabels); err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %w", err)
	}

	return metrics, nil
}

// splitLabels splits label string on commas, but respects quoted values
func (s *MetricsService) splitLabels(labelStr string) []string {
	var labels []string
	var current strings.Builder
	inQuotes := false

	for _, r := range labelStr {
		switch r {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case ',':
			if !inQuotes {
				if current.Len() > 0 {
					labels = append(labels, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		labels = append(labels, current.String())
	}

	return labels
}

// parseAllMetrics parses all required metrics
func (s *MetricsService) parseAllMetrics(metrics *models.SystemMetrics, metricValues map[string]float64, metricLabels map[string]map[string]string) error {
	// Parse CPU metrics
	s.parseCPUMetrics(metrics, metricValues, metricLabels)

	// Parse Memory metrics
	s.parseMemoryMetrics(metrics, metricValues)

	// Parse Disk metrics
	s.parseDiskMetrics(metrics, metricValues, metricLabels)

	return nil
}

// parseCPUMetrics calculates CPU usage using delta approach (like Grafana irate)
func (s *MetricsService) parseCPUMetrics(metrics *models.SystemMetrics, metricValues map[string]float64, metricLabels map[string]map[string]string) {
	// Count CPU cores
	cpuCores := make(map[string]bool)
	currentCPUMetrics := make(map[string]float64)

	for key, value := range metricValues {
		if strings.HasPrefix(key, "node_cpu_seconds_total") {
			labels := metricLabels[key]
			cpu := labels["cpu"]
			mode := labels["mode"]

			if cpu != "" && mode != "" {
				cpuCores[cpu] = true
				// Store current metrics for delta calculation
				cpuKey := cpu + "_" + mode
				currentCPUMetrics[cpuKey] = value
			}
		}
	}

	// Set allocated CPU
	metrics.AllocatedCPU = int64(len(cpuCores))

	s.cpuHistory.mu.Lock()
	defer s.cpuHistory.mu.Unlock()

	now := time.Now()

	// If we have previous data, calculate deltas
	if len(s.cpuHistory.prevMetrics) > 0 && !s.cpuHistory.lastUpdate.IsZero() {
		timeDelta := now.Sub(s.cpuHistory.lastUpdate).Seconds()

		if timeDelta > 0 && timeDelta < 300 { // Only use if reasonable time interval (< 5 minutes)
			var totalIdleDelta, totalAllDelta float64

			for cpu := range cpuCores {
				var cpuTotalDelta, cpuIdleDelta float64

				// Calculate deltas for each CPU mode
				for _, mode := range []string{"idle", "user", "system", "iowait", "irq", "softirq", "steal", "guest", "guest_nice"} {
					cpuKey := cpu + "_" + mode

					if currentVal, exists := currentCPUMetrics[cpuKey]; exists {
						if prevVal, hasPrev := s.cpuHistory.prevMetrics[cpuKey]; hasPrev {
							delta := currentVal - prevVal
							if delta >= 0 { // Ensure positive delta
								cpuTotalDelta += delta
								if mode == "idle" {
									cpuIdleDelta = delta
								}
							}
						}
					}
				}

				totalAllDelta += cpuTotalDelta
				totalIdleDelta += cpuIdleDelta
			}

			// Calculate CPU usage: (1 - idle_rate) * 100
			if totalAllDelta > 0 {
				idlePercent := totalIdleDelta / totalAllDelta
				metrics.CPUUsage = (1 - idlePercent) * 100

				// Ensure reasonable bounds
				if metrics.CPUUsage < 0 {
					metrics.CPUUsage = 0
				} else if metrics.CPUUsage > 100 {
					metrics.CPUUsage = 100
				}
			}
		}
	}

	// Fallback to load average if delta calculation didn't work
	if metrics.CPUUsage == 0 {
		if load1, exists := metricValues["node_load1"]; exists && metrics.AllocatedCPU > 0 {
			loadPercent := (load1 / float64(metrics.AllocatedCPU)) * 100
			if loadPercent > 100 {
				loadPercent = 100
			}
			metrics.CPUUsage = loadPercent
		}
	}

	// Store current metrics for next iteration
	s.cpuHistory.prevMetrics = currentCPUMetrics
	s.cpuHistory.lastUpdate = now
}

// parseMemoryMetrics calculates RAM usage and allocated memory using Grafana-style formula
func (s *MetricsService) parseMemoryMetrics(metrics *models.SystemMetrics, metricValues map[string]float64) {
	// Grafana formula: sum(node_memory_MemTotal_bytes)
	totalMemory := metricValues["node_memory_MemTotal_bytes"]

	// Grafana formula components for memory usage:
	// 100 * (1 - ((node_memory_MemFree_bytes + node_memory_Cached_bytes + node_memory_Buffers_bytes) / node_memory_MemTotal_bytes))
	memFree := metricValues["node_memory_MemFree_bytes"]
	memCached := metricValues["node_memory_Cached_bytes"]
	memBuffers := metricValues["node_memory_Buffers_bytes"]

	if totalMemory > 0 {
		// Convert bytes to gigabytes (1 GB = 1024^3 bytes)
		metrics.AllocatedMemory = int64(totalMemory / (1024 * 1024 * 1024))

		// Calculate memory usage using Grafana formula
		availableMemory := memFree + memCached + memBuffers
		metrics.RAMUsage = 100 * (1 - (availableMemory / totalMemory))

		// Ensure percentage is within valid range
		if metrics.RAMUsage < 0 {
			metrics.RAMUsage = 0
		} else if metrics.RAMUsage > 100 {
			metrics.RAMUsage = 100
		}
	}
}

// parseDiskMetrics calculates disk usage and allocated disk space using Grafana-style formula
func (s *MetricsService) parseDiskMetrics(metrics *models.SystemMetrics, metricValues map[string]float64, metricLabels map[string]map[string]string) {
	// Grafana queries focus on specific filesystem types and mount points
	// We'll look for xfs filesystems and common mount points like /var/lib/docker, /, etc.

	var totalDiskSize, totalDiskFree float64
	validDevices := 0

	// Parse filesystem metrics - focusing on main filesystems
	// Grafana query: node_filesystem_size_bytes{fstype=~"xfs",mountpoint=~"/var/lib/docker"}
	// We'll be more flexible and include common mount points and filesystem types

	for key, value := range metricValues {
		labels := metricLabels[key]
		device := labels["device"]
		mountpoint := labels["mountpoint"]
		fstype := labels["fstype"]

		if strings.HasPrefix(key, "node_filesystem_size_bytes") && s.isValidDiskDevice(device, mountpoint, fstype) {
			totalDiskSize += value
			validDevices++
		}

		if strings.HasPrefix(key, "node_filesystem_free_bytes") && s.isValidDiskDevice(device, mountpoint, fstype) {
			totalDiskFree += value
		}
	}

	if totalDiskSize > 0 {
		// Convert bytes to gigabytes (1 GB = 1024^3 bytes)
		metrics.AllocatedDisk = int64(totalDiskSize / (1024 * 1024 * 1024))

		// Calculate disk usage using Grafana formula:
		// 100*(1-(node_filesystem_free_bytes / node_filesystem_size_bytes))
		metrics.DiskUsage = 100 * (1 - (totalDiskFree / totalDiskSize))

		// Ensure percentage is within valid range
		if metrics.DiskUsage < 0 {
			metrics.DiskUsage = 0
		} else if metrics.DiskUsage > 100 {
			metrics.DiskUsage = 100
		}
	}
}

// isValidDiskDevice determines if a device should be included in disk calculations
// Based on Grafana query patterns: fstype=~"xfs", physical devices, important mount points
func (s *MetricsService) isValidDiskDevice(device, mountpoint, fstype string) bool {
	// Skip non-physical devices
	if !strings.HasPrefix(device, "/dev/") {
		return false
	}

	// Skip loop devices and snap mounts
	if strings.Contains(device, "loop") || strings.Contains(mountpoint, "snap") {
		return false
	}

	// Include common filesystem types (xfs, ext4, ext3, etc.)
	validFSTypes := []string{"xfs", "ext4", "ext3", "ext2", "btrfs"}
	validFS := false
	for _, validType := range validFSTypes {
		if fstype == validType {
			validFS = true
			break
		}
	}
	if !validFS {
		return false
	}

	// Include important mount points or root filesystem
	importantMounts := []string{"/", "/var/lib/docker", "/home", "/opt", "/usr", "/var"}
	for _, mount := range importantMounts {
		if mountpoint == mount || strings.HasPrefix(mountpoint, mount+"/") {
			return true
		}
	}

	// Include if it's a root mount or major partition
	if mountpoint == "/" || (mountpoint != "" && !strings.Contains(mountpoint, "boot")) {
		return true
	}

	return false
}

// GetSystemMetrics fetches and parses system metrics
func (s *MetricsService) GetSystemMetrics(ctx context.Context) (*models.SystemMetrics, error) {
	metricsText, err := s.fetchMetrics()
	if err != nil {
		return nil, err
	}

	return s.parseMetrics(metricsText)
}

func (s *MetricsService) GetSnapshotCount(ctx context.Context) (int, error) {
	images, err := s.docker.ApiClient().ImageList(ctx, image.ListOptions{})
	if err != nil {
		return -1, nil // Return -1 on error
	}

	return len(images), nil
}

// GetCachedSystemMetrics returns cached metrics if available, otherwise returns defaults
func (s *MetricsService) GetCachedSystemMetrics(ctx context.Context) *models.SystemMetrics {
	metrics, err := s.cache.Get(ctx, systemMetricsKey)
	if err != nil || metrics == nil {
		// Return default values if no metrics are cached
		return &models.SystemMetrics{
			CPUUsage:        -1.0,
			RAMUsage:        -1.0,
			DiskUsage:       -1.0,
			AllocatedCPU:    -1,
			AllocatedMemory: -1,
			AllocatedDisk:   -1,
			SnapshotCount:   -1,
			LastUpdated:     time.Now(),
		}
	}

	return metrics
}

// CollectAndCacheMetrics collects all metrics and stores them in cache
func (s *MetricsService) CollectAndCacheMetrics(ctx context.Context) error {
	finalMetrics := &models.SystemMetrics{
		CPUUsage:        -1.0,
		RAMUsage:        -1.0,
		DiskUsage:       -1.0,
		AllocatedCPU:    -1,
		AllocatedMemory: -1,
		AllocatedDisk:   -1,
		SnapshotCount:   -1,
		LastUpdated:     time.Now(),
	}

	// Get current cached metrics to preserve valid values
	cachedMetrics, err := s.cache.Get(ctx, systemMetricsKey)
	if err == nil && cachedMetrics != nil {
		finalMetrics = cachedMetrics
	}

	// Collect system metrics
	metrics, err := s.GetSystemMetrics(ctx)
	if err == nil {
		// Only update if values are valid (not 0 or -1)
		if metrics.CPUUsage > 0 && metrics.CPUUsage != -1.0 {
			finalMetrics.CPUUsage = metrics.CPUUsage
		}
		if metrics.RAMUsage > 0 && metrics.RAMUsage != -1.0 {
			finalMetrics.RAMUsage = metrics.RAMUsage
		}
		if metrics.DiskUsage > 0 && metrics.DiskUsage != -1.0 {
			finalMetrics.DiskUsage = metrics.DiskUsage
		}
		if metrics.AllocatedCPU > 0 && metrics.AllocatedCPU != -1 {
			finalMetrics.AllocatedCPU = metrics.AllocatedCPU
		}
		if metrics.AllocatedMemory > 0 && metrics.AllocatedMemory != -1 {
			finalMetrics.AllocatedMemory = metrics.AllocatedMemory
		}
		if metrics.AllocatedDisk > 0 && metrics.AllocatedDisk != -1 {
			finalMetrics.AllocatedDisk = metrics.AllocatedDisk
		}
	}

	// Get snapshot count
	snapshotCount, err := s.GetSnapshotCount(ctx)
	if err == nil && snapshotCount > 0 && snapshotCount != -1 {
		finalMetrics.SnapshotCount = snapshotCount
	}

	// Store in cache with final values
	s.cache.Set(ctx, systemMetricsKey, *finalMetrics, 2*time.Hour)

	return nil
}

// StartMetricsCollection starts a background goroutine that collects metrics every 20 seconds
func (s *MetricsService) StartMetricsCollection(ctx context.Context) {
	go func() {
		// Collect metrics immediately on startup
		_ = s.CollectAndCacheMetrics(ctx)

		// Set up ticker for every 20 seconds
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = s.CollectAndCacheMetrics(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}
