// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package loadstat reads allocation-relative resource load from inside a sandbox.
//
// The whole point is to be cgroup-correct: inside a container, /proc, `nproc` and
// `free` report the HOST's CPU/memory, not the sandbox's allocation. The only
// trustworthy, allocation-relative signals are the cgroup files themselves plus PSI
// (Pressure Stall Information). We prefer cgroup v2, fall back to v1, and degrade
// gracefully (returning nil sub-blocks) when neither is readable — the API then
// falls back to concurrency-only saturation.
package loadstat

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func numCPU() int { return runtime.NumCPU() }

// CPU load. Utilization is fraction (0..1) of the cgroup's CPU quota used since the
// previous sample; PressureSomeAvg10 is the PSI "some avg10" percentage (0..100).
type CPU struct {
	Utilization       float64 `json:"utilization,omitempty"`
	PressureSomeAvg10 float64 `json:"pressureSomeAvg10,omitempty"`
}

// Memory load. Utilization is current/limit (0..1).
type Memory struct {
	Utilization       float64 `json:"utilization,omitempty"`
	PressureSomeAvg10 float64 `json:"pressureSomeAvg10,omitempty"`
}

// IO load — only PSI is meaningful for "under load" detection.
type IO struct {
	PressureSomeAvg10 float64 `json:"pressureSomeAvg10,omitempty"`
}

// Disk load for the workspace volume. Utilization is used/total (0..1).
type Disk struct {
	Utilization float64 `json:"utilization,omitempty"`
}

// Sample is the full resource snapshot. Any sub-block may be nil when unreadable.
type Sample struct {
	CPU    *CPU    `json:"cpu,omitempty"`
	Memory *Memory `json:"memory,omitempty"`
	IO     *IO     `json:"io,omitempty"`
	Disk   *Disk   `json:"disk,omitempty"`
}

// Collector reads cgroup + statfs metrics. It is stateful only for CPU utilization,
// which is a delta of cgroup CPU usage over wall-clock time between samples.
type Collector struct {
	cgroupRoot    string
	workspaceRoot string

	mu          sync.Mutex
	prevCPUUsec uint64
	prevAt      time.Time
}

// NewCollector builds a collector. cgroupRoot defaults to /sys/fs/cgroup.
func NewCollector(cgroupRoot, workspaceRoot string) *Collector {
	if cgroupRoot == "" {
		cgroupRoot = "/sys/fs/cgroup"
	}
	return &Collector{cgroupRoot: cgroupRoot, workspaceRoot: workspaceRoot}
}

// Sample reads the current resource snapshot. now is injectable for tests.
func (c *Collector) Sample(now time.Time) *Sample {
	s := &Sample{}
	if c.isV2() {
		s.CPU = c.cpuV2(now)
		s.Memory = c.memV2()
		s.IO = c.ioV2()
	} else {
		s.CPU = c.cpuV1(now)
		s.Memory = c.memV1()
	}
	s.Disk = c.disk()
	return s
}

func (c *Collector) isV2() bool {
	// Unified hierarchy exposes cgroup.controllers at the root.
	_, err := os.Stat(filepath.Join(c.cgroupRoot, "cgroup.controllers"))
	return err == nil
}

func (c *Collector) read(rel string) (string, bool) {
	b, err := os.ReadFile(filepath.Join(c.cgroupRoot, rel))
	if err != nil {
		return "", false
	}
	return string(b), true
}

// -- cgroup v2 -----------------------------------------------------------

func (c *Collector) cpuV2(now time.Time) *CPU {
	out := &CPU{}
	// Track whether ANY real CPU signal was obtained (PSI or usage). A valid PSI
	// of 0 is a real reading, so we must not infer "missing" from the value being
	// zero — only from the parse failing. Return nil only when nothing was
	// readable, matching the graceful-degradation contract of cpuV1/memV1/etc.
	got := false
	if raw, ok := c.read("cpu.pressure"); ok {
		if v, ok := ParsePSISomeAvg10(raw); ok {
			out.PressureSomeAvg10 = v
			got = true
		}
	}
	statRaw, ok := c.read("cpu.stat")
	if !ok {
		if !got {
			return nil
		}
		return out
	}
	usage, ok := ParseCPUUsageUsec(statRaw)
	if !ok {
		// cpu.stat readable but malformed: fall back to whatever PSI we got, or
		// nil if we got nothing at all (don't return an empty &CPU{}).
		if !got {
			return nil
		}
		return out
	}
	quota := 0.0
	if maxRaw, ok := c.read("cpu.max"); ok {
		if q, period, ok := ParseCPUMax(maxRaw); ok && q > 0 && period > 0 {
			quota = q / period // cores
		}
	}
	if quota <= 0 {
		// "max" (unlimited) — normalize against host CPU count as a best effort.
		quota = float64(numCPU())
	}
	out.Utilization = c.cpuUtil(usage, now, quota)
	return out
}

// cpuUtil computes fraction-of-quota used since the previous sample.
func (c *Collector) cpuUtil(usageUsec uint64, now time.Time, quotaCores float64) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	prevUsec, prevAt := c.prevCPUUsec, c.prevAt
	c.prevCPUUsec, c.prevAt = usageUsec, now
	if prevAt.IsZero() || !now.After(prevAt) || usageUsec < prevUsec || quotaCores <= 0 {
		return 0
	}
	elapsedUsec := float64(now.Sub(prevAt).Microseconds())
	if elapsedUsec <= 0 {
		return 0
	}
	used := float64(usageUsec-prevUsec) / (elapsedUsec * quotaCores)
	return clamp01(used)
}

func (c *Collector) memV2() *Memory {
	out := &Memory{}
	got := false
	if raw, ok := c.read("memory.pressure"); ok {
		if v, ok := ParsePSISomeAvg10(raw); ok {
			out.PressureSomeAvg10 = v
			got = true
		}
	}
	cur, okC := c.read("memory.current")
	max, okM := c.read("memory.max")
	if okC && okM {
		if u, ok := ParseMemUtil(cur, max); ok {
			out.Utilization = u
			got = true
		}
	}
	if !got {
		return nil
	}
	return out
}

func (c *Collector) ioV2() *IO {
	if raw, ok := c.read("io.pressure"); ok {
		if v, ok := ParsePSISomeAvg10(raw); ok {
			return &IO{PressureSomeAvg10: v}
		}
	}
	return nil
}

// -- cgroup v1 (no PSI) --------------------------------------------------

func (c *Collector) cpuV1(now time.Time) *CPU {
	// v1 cpuacct usage is in nanoseconds; try the common mount layouts.
	for _, rel := range []string{"cpu,cpuacct/cpuacct.usage", "cpuacct/cpuacct.usage", "cpuacct.usage"} {
		if raw, ok := c.read(rel); ok {
			if ns, ok := ParseUint(strings.TrimSpace(raw)); ok {
				usageUsec := ns / 1000
				return &CPU{Utilization: c.cpuUtil(usageUsec, now, float64(numCPU()))}
			}
		}
	}
	return nil
}

func (c *Collector) memV1() *Memory {
	var usage, limit uint64
	var okU, okL bool
	for _, rel := range []string{"memory/memory.usage_in_bytes", "memory.usage_in_bytes"} {
		if raw, ok := c.read(rel); ok {
			usage, okU = ParseUint(strings.TrimSpace(raw))
			break
		}
	}
	for _, rel := range []string{"memory/memory.limit_in_bytes", "memory.limit_in_bytes"} {
		if raw, ok := c.read(rel); ok {
			limit, okL = ParseUint(strings.TrimSpace(raw))
			break
		}
	}
	// v1 reports an enormous sentinel when unlimited — treat as no signal.
	if okU && okL && limit > 0 && limit < (1<<62) {
		return &Memory{Utilization: clamp01(float64(usage) / float64(limit))}
	}
	return nil
}

// -- disk ----------------------------------------------------------------

func (c *Collector) disk() *Disk {
	if c.workspaceRoot == "" {
		return nil
	}
	var st syscall.Statfs_t
	if err := syscall.Statfs(c.workspaceRoot, &st); err != nil {
		return nil
	}
	total := st.Blocks
	if total == 0 {
		return nil
	}
	used := total - st.Bfree
	return &Disk{Utilization: clamp01(float64(used) / float64(total))}
}

// -- pure parsers (unit-tested) ------------------------------------------

// ParsePSISomeAvg10 extracts the "some avg10=N" value from a PSI file body.
func ParsePSISomeAvg10(content string) (float64, bool) {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "some" {
			continue
		}
		for _, f := range fields[1:] {
			if strings.HasPrefix(f, "avg10=") {
				v, err := strconv.ParseFloat(strings.TrimPrefix(f, "avg10="), 64)
				if err != nil {
					return 0, false
				}
				return v, true
			}
		}
	}
	return 0, false
}

// ParseCPUUsageUsec extracts "usage_usec N" from a cgroup v2 cpu.stat body.
func ParseCPUUsageUsec(content string) (uint64, bool) {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "usage_usec" {
			return ParseUint(fields[1])
		}
	}
	return 0, false
}

// ParseCPUMax parses a cgroup v2 cpu.max body ("<quota> <period>" or "max <period>").
// Returns quota and period in microseconds. When quota is "max" it returns ok=false
// for the quota (callers treat that as unlimited).
func ParseCPUMax(content string) (quotaUsec, periodUsec float64, ok bool) {
	fields := strings.Fields(strings.TrimSpace(content))
	if len(fields) != 2 {
		return 0, 0, false
	}
	if fields[0] == "max" {
		return 0, 0, false
	}
	q, err1 := strconv.ParseFloat(fields[0], 64)
	p, err2 := strconv.ParseFloat(fields[1], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return q, p, true
}

// ParseMemUtil computes current/max from cgroup v2 memory.current and memory.max
// bodies. memory.max of "max" (unlimited) yields ok=false.
func ParseMemUtil(currentRaw, maxRaw string) (float64, bool) {
	cur, ok := ParseUint(strings.TrimSpace(currentRaw))
	if !ok {
		return 0, false
	}
	maxStr := strings.TrimSpace(maxRaw)
	if maxStr == "max" {
		return 0, false
	}
	max, ok := ParseUint(maxStr)
	if !ok || max == 0 {
		return 0, false
	}
	return clamp01(float64(cur) / float64(max)), true
}

// ParseUint parses an unsigned base-10 integer.
func ParseUint(s string) (uint64, bool) {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
