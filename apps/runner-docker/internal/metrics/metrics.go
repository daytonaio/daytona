// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusOperationStatus string

const (
	PrometheusOperationStatusSuccess PrometheusOperationStatus = "success"
	PrometheusOperationStatusFailure PrometheusOperationStatus = "failure"

	CreateSandboxOperation  = "create"
	DestroySandboxOperation = "destroy"
)

// Define your metrics
var (
	// Histogram to track duration of container operations
	ContainerOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "container_operation_duration_seconds",
			Help: "Time taken for container operations in seconds",
			// Buckets optimized for detecting anomalies in operation durations
			Buckets: []float64{0.1, 0.25, 0.5, 0.75, 1, 2, 3, 5, 7.5, 10, 15, 30, 60, 120, 300},
		},
		[]string{"operation"},
	)

	// Counter to track occurrence of container operations with status
	ContainerOperationCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "container_operation_total",
			Help: "Total number of container operations",
		},
		[]string{"operation", "status"},
	)
)

func SuccessCounterInc(operation string) {
	ContainerOperationCount.WithLabelValues(operation, string(PrometheusOperationStatusSuccess)).Inc()
}

func FailureCounterInc(operation string) {
	ContainerOperationCount.WithLabelValues(operation, string(PrometheusOperationStatusFailure)).Inc()
}
