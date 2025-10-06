// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package metrics

// MockCollector implements MetricsCollector with static mock values
type MockCollector struct{}

// NewMockCollector creates a new instance of MockCollector
func NewMockCollector() *MockCollector {
	return &MockCollector{}
}

func (m *MockCollector) GetCPUPercentage() (float64, error) {
	return 50.0, nil
}

func (m *MockCollector) GetMemoryPercentage() (float64, error) {
	return 50.0, nil
}

func (m *MockCollector) GetDiskPercentage() (float64, error) {
	return 50.0, nil
}
