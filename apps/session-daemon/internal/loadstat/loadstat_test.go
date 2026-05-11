// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package loadstat

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParsePSISomeAvg10(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		want   float64
		wantOK bool
	}{
		{
			name:   "typical",
			in:     "some avg10=12.34 avg60=5.00 avg300=1.00 total=123456\nfull avg10=1.00 avg60=0.00 avg300=0.00 total=10",
			want:   12.34,
			wantOK: true,
		},
		{name: "zero", in: "some avg10=0.00 avg60=0.00 avg300=0.00 total=0", want: 0, wantOK: true},
		{name: "no some line", in: "full avg10=1.00 total=10", want: 0, wantOK: false},
		{name: "empty", in: "", want: 0, wantOK: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParsePSISomeAvg10(tc.in)
			if ok != tc.wantOK || (ok && got != tc.want) {
				t.Fatalf("ParsePSISomeAvg10(%q) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestParseCPUUsageUsec(t *testing.T) {
	in := "usage_usec 123456789\nuser_usec 100000000\nsystem_usec 23456789\nnr_periods 0"
	got, ok := ParseCPUUsageUsec(in)
	if !ok || got != 123456789 {
		t.Fatalf("ParseCPUUsageUsec = (%d, %v), want (123456789, true)", got, ok)
	}
	if _, ok := ParseCPUUsageUsec("user_usec 1\n"); ok {
		t.Fatalf("expected ok=false when usage_usec absent")
	}
}

func TestParseCPUMax(t *testing.T) {
	q, p, ok := ParseCPUMax("100000 100000")
	if !ok || q != 100000 || p != 100000 {
		t.Fatalf("ParseCPUMax(quota) = (%v,%v,%v)", q, p, ok)
	}
	q, p, ok = ParseCPUMax("200000 100000\n")
	if !ok || q/p != 2 {
		t.Fatalf("ParseCPUMax(2 cores) = (%v,%v,%v)", q, p, ok)
	}
	if _, _, ok := ParseCPUMax("max 100000"); ok {
		t.Fatalf("ParseCPUMax(max) should be ok=false (unlimited)")
	}
	if _, _, ok := ParseCPUMax("garbage"); ok {
		t.Fatalf("ParseCPUMax(garbage) should be ok=false")
	}
}

func TestParseMemUtil(t *testing.T) {
	u, ok := ParseMemUtil("500", "1000")
	if !ok || u != 0.5 {
		t.Fatalf("ParseMemUtil(500/1000) = (%v,%v), want (0.5,true)", u, ok)
	}
	if _, ok := ParseMemUtil("500", "max"); ok {
		t.Fatalf("ParseMemUtil(max) should be ok=false")
	}
	// over-limit clamps to 1.
	if u, ok := ParseMemUtil("2000", "1000"); !ok || u != 1 {
		t.Fatalf("ParseMemUtil over-limit = (%v,%v), want (1,true)", u, ok)
	}
}

// TestCollectorV2 builds a fake cgroup v2 tree and checks the composed sample.
func TestCollectorV2(t *testing.T) {
	root := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("cgroup.controllers", "cpu memory io")
	write("cpu.max", "200000 100000") // 2 cores
	write("cpu.stat", "usage_usec 1000000\n")
	write("cpu.pressure", "some avg10=7.50 avg60=1.00 total=1\nfull avg10=0.00 total=0")
	write("memory.current", "500")
	write("memory.max", "1000")
	write("memory.pressure", "some avg10=3.20 total=1")
	write("io.pressure", "some avg10=1.10 total=1")

	c := NewCollector(root, root)
	// First sample seeds the CPU delta baseline (utilization 0).
	t0 := time.Unix(1000, 0)
	s0 := c.Sample(t0)
	if s0.CPU == nil || s0.CPU.PressureSomeAvg10 != 7.5 {
		t.Fatalf("first sample cpu pressure = %+v", s0.CPU)
	}
	if s0.Memory == nil || s0.Memory.Utilization != 0.5 || s0.Memory.PressureSomeAvg10 != 3.2 {
		t.Fatalf("memory = %+v", s0.Memory)
	}
	if s0.IO == nil || s0.IO.PressureSomeAvg10 != 1.1 {
		t.Fatalf("io = %+v", s0.IO)
	}
	if s0.Disk == nil {
		t.Fatalf("disk should be present (statfs of tempdir)")
	}

	// Advance usage by 1 core-second over 1 wall-second with a 2-core quota => 0.5 util.
	write("cpu.stat", "usage_usec 2000000\n")
	s1 := c.Sample(t0.Add(time.Second))
	if s1.CPU == nil || s1.CPU.Utilization < 0.49 || s1.CPU.Utilization > 0.51 {
		t.Fatalf("cpu utilization = %+v, want ~0.5", s1.CPU)
	}
}

func TestCollectorNoCgroupDegradesGracefully(t *testing.T) {
	root := t.TempDir() // empty: no cgroup.controllers, no v1 files
	c := NewCollector(root, root)
	s := c.Sample(time.Now())
	if s.CPU != nil || s.Memory != nil || s.IO != nil {
		t.Fatalf("expected nil resource blocks with no cgroup files, got %+v", s)
	}
	// disk still works via statfs.
	if s.Disk == nil {
		t.Fatalf("disk should still be readable")
	}
}
