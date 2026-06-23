// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"math"
	"testing"
)

func TestCPUUsagePercent(t *testing.T) {
	const sec = int64(1_000_000_000) // 1 second in nanoseconds

	tests := []struct {
		name      string
		cpuDelta  uint64
		wallDelta int64
		cpuLimit  float64
		want      float64
	}{
		{"half of one core", uint64(sec / 2), sec, 1.0, 50.0},
		{"one core fully loaded", uint64(sec), sec, 1.0, 100.0},
		{"two cores fully loaded", uint64(2 * sec), sec, 2.0, 100.0},
		{"one of two cores busy", uint64(sec), sec, 2.0, 50.0},
		{"half-core limit fully loaded", uint64(sec / 2), sec, 0.5, 100.0},
		{"idle", 0, sec, 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CPUUsagePercent(tt.cpuDelta, tt.wallDelta, tt.cpuLimit)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("CPUUsagePercent(%d, %d, %g) = %g, want %g",
					tt.cpuDelta, tt.wallDelta, tt.cpuLimit, got, tt.want)
			}
		})
	}
}
