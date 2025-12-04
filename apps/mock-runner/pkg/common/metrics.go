// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusOperationStatus string

const (
	PrometheusOperationStatusSuccess PrometheusOperationStatus = "success"
	PrometheusOperationStatusFailure PrometheusOperationStatus = "failure"
)

var (
	ContainerOperationCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mock_runner_container_operation_total",
			Help: "Total number of container operations",
		},
		[]string{"operation", "status"},
	)

	ContainerOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mock_runner_container_operation_duration_seconds",
			Help:    "Duration of container operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)



