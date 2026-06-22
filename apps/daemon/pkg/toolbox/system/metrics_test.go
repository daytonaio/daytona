// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package system

import (
	"math"
	"testing"
	"time"
)

func TestSamplerUpdate(t *testing.T) {
	const sec = int64(1_000_000_000)
	t0 := time.Unix(1_700_000_000, 0)

	t.Run("cold start leaves cpuUsedPct at 0", func(t *testing.T) {
		s := &sampler{}
		s.update(1000, t0, 1.0)
		if s.cpuUsedPct != 0 {
			t.Errorf("cold-start cpuUsedPct = %g, want 0", s.cpuUsedPct)
		}
	})

	t.Run("normal delta computes percentage", func(t *testing.T) {
		s := &sampler{}
		s.update(0, t0, 1.0)
		s.update(uint64(sec/2), t0.Add(time.Second), 1.0)
		if math.Abs(s.cpuUsedPct-50.0) > 0.001 {
			t.Errorf("cpuUsedPct = %g, want 50", s.cpuUsedPct)
		}
	})

	t.Run("counter reset zeros cpuUsedPct", func(t *testing.T) {
		s := &sampler{}
		s.update(0, t0, 1.0)
		s.update(uint64(sec), t0.Add(time.Second), 1.0)
		if s.cpuUsedPct == 0 {
			t.Fatal("precondition: expected non-zero before reset")
		}
		s.update(10, t0.Add(2*time.Second), 1.0) // counter went backwards
		if s.cpuUsedPct != 0 {
			t.Errorf("after reset cpuUsedPct = %g, want 0", s.cpuUsedPct)
		}
	})

	t.Run("idle (unchanged counter) yields 0", func(t *testing.T) {
		s := &sampler{}
		s.update(5000, t0, 1.0)
		s.update(5000, t0.Add(time.Second), 1.0)
		if s.cpuUsedPct != 0 {
			t.Errorf("idle cpuUsedPct = %g, want 0", s.cpuUsedPct)
		}
	})

	t.Run("non-monotonic clock zeros cpuUsedPct", func(t *testing.T) {
		s := &sampler{}
		s.update(0, t0, 1.0)
		s.update(uint64(sec), t0.Add(time.Second), 1.0)
		s.update(uint64(2*sec), t0, 1.0) // clock jumped backwards
		if s.cpuUsedPct != 0 {
			t.Errorf("backwards-clock cpuUsedPct = %g, want 0", s.cpuUsedPct)
		}
	})

	t.Run("zero cpu limit never computes", func(t *testing.T) {
		s := &sampler{}
		s.update(0, t0, 0)
		s.update(uint64(sec), t0.Add(time.Second), 0)
		if s.cpuUsedPct != 0 {
			t.Errorf("zero-limit cpuUsedPct = %g, want 0", s.cpuUsedPct)
		}
	})
}
