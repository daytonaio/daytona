// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package metrics

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/procfs"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeProcEntry fakes /proc/<pid> with a cgroup file, a procfs-format limits
// file, and openFds dummy entries under fd/.
func writeProcEntry(t *testing.T, procDir string, pid int, cgroupLine string, openFds int, softFdLimit string) {
	t.Helper()

	pidDir := filepath.Join(procDir, strconv.Itoa(pid))
	writeFile(t, filepath.Join(pidDir, "cgroup"), cgroupLine+"\n")
	writeFile(t, filepath.Join(pidDir, "limits"),
		"Limit                     Soft Limit           Hard Limit           Units\n"+
			fmt.Sprintf("Max open files            %-21s%-21sfiles\n", softFdLimit, softFdLimit))

	for i := range openFds {
		writeFile(t, filepath.Join(pidDir, "fd", strconv.Itoa(i)), "")
	}
	if openFds == 0 {
		if err := os.MkdirAll(filepath.Join(pidDir, "fd"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSampleSandboxFdUsage(t *testing.T) {
	root := t.TempDir()
	procDir := filepath.Join(root, "proc")
	cgroupDir := filepath.Join(root, "cgroup")

	writeProcEntry(t, procDir, 100, "0::/sandbox-test", 3, "100")
	writeProcEntry(t, procDir, 101, "0::/sandbox-test", 8, "10")
	writeProcEntry(t, procDir, 102, "0::/sandbox-test", 5, "unlimited")
	// PID 999 is listed in cgroup.procs but missing from /proc: exited
	// between the cgroup.procs read and sampling, must be skipped.
	writeFile(t, filepath.Join(cgroupDir, "sandbox-test", "cgroup.procs"), "100\n101\n102\n999\n")

	usage, err := sampleSandboxFdUsage(procDir, cgroupDir, 100)
	if err != nil {
		t.Fatalf("sampleSandboxFdUsage: %v", err)
	}

	if usage.OpenFds != 16 {
		t.Errorf("OpenFds = %d, want 16 (3+8+5, unlimited counts toward the sum)", usage.OpenFds)
	}
	if usage.UsagePercent != 80 {
		t.Errorf("UsagePercent = %v, want 80 (worst process 101: 8/10)", usage.UsagePercent)
	}
	if usage.WorstPid != 101 {
		t.Errorf("WorstPid = %d, want 101", usage.WorstPid)
	}
	if usage.WorstOpenFds != 8 {
		t.Errorf("WorstOpenFds = %d, want 8", usage.WorstOpenFds)
	}
	if usage.WorstFdLimit != 10 {
		t.Errorf("WorstFdLimit = %d, want 10", usage.WorstFdLimit)
	}
}

func TestSampleSandboxFdUsageAllUnlimited(t *testing.T) {
	root := t.TempDir()
	procDir := filepath.Join(root, "proc")
	cgroupDir := filepath.Join(root, "cgroup")

	writeProcEntry(t, procDir, 100, "0::/sandbox-test", 4, "unlimited")
	writeFile(t, filepath.Join(cgroupDir, "sandbox-test", "cgroup.procs"), "100\n")

	usage, err := sampleSandboxFdUsage(procDir, cgroupDir, 100)
	if err != nil {
		t.Fatalf("sampleSandboxFdUsage: %v", err)
	}

	if usage.OpenFds != 4 {
		t.Errorf("OpenFds = %d, want 4", usage.OpenFds)
	}
	if usage.UsagePercent != 0 {
		t.Errorf("UsagePercent = %v, want 0 for unlimited-only processes", usage.UsagePercent)
	}
	if usage.WorstPid != 0 {
		t.Errorf("WorstPid = %d, want 0 when no process has a finite limit", usage.WorstPid)
	}
}

func TestResolveCgroupDir(t *testing.T) {
	cgroupRoot := "/sys/fs/cgroup"

	t.Run("v2 unified", func(t *testing.T) {
		dir, err := resolveCgroupDir([]procfs.Cgroup{
			{HierarchyID: 0, Path: "/system.slice/docker-x.scope"},
		}, cgroupRoot)
		if err != nil {
			t.Fatalf("resolveCgroupDir: %v", err)
		}
		if want := "/sys/fs/cgroup/system.slice/docker-x.scope"; dir != want {
			t.Errorf("dir = %q, want %q", dir, want)
		}
	})

	t.Run("v1 pids controller", func(t *testing.T) {
		dir, err := resolveCgroupDir([]procfs.Cgroup{
			{HierarchyID: 3, Controllers: []string{"cpu", "cpuacct"}, Path: "/docker/abc"},
			{HierarchyID: 5, Controllers: []string{"pids"}, Path: "/docker/abc"},
		}, cgroupRoot)
		if err != nil {
			t.Fatalf("resolveCgroupDir: %v", err)
		}
		if want := "/sys/fs/cgroup/pids/docker/abc"; dir != want {
			t.Errorf("dir = %q, want %q", dir, want)
		}
	})

	t.Run("no usable hierarchy", func(t *testing.T) {
		if _, err := resolveCgroupDir([]procfs.Cgroup{
			{HierarchyID: 3, Controllers: []string{"cpu"}, Path: "/docker/abc"},
		}, cgroupRoot); err == nil {
			t.Error("expected error for entries without unified or pids hierarchy")
		}
	})
}

func TestFdWarnTrackerHysteresis(t *testing.T) {
	tracker := newFdWarnTracker(70)
	now := time.Now()

	if got := tracker.observe("sb", 70, now); got != fdWarnEventWarn {
		t.Fatalf("crossing threshold: got %v, want warn", got)
	}
	if got := tracker.observe("sb", 67, now.Add(time.Second)); got != fdWarnEventNone {
		t.Errorf("inside hysteresis band (65-70): got %v, want none", got)
	}
	if got := tracker.observe("sb", 71, now.Add(2*time.Second)); got != fdWarnEventNone {
		t.Errorf("re-crossing threshold while active inside re-warn window: got %v, want none", got)
	}
	if got := tracker.observe("sb", 64, now.Add(3*time.Second)); got != fdWarnEventClear {
		t.Errorf("dropping below band: got %v, want clear", got)
	}
	if got := tracker.observe("sb", 64, now.Add(4*time.Second)); got != fdWarnEventNone {
		t.Errorf("staying below band after clear: got %v, want none", got)
	}
	if got := tracker.observe("sb", 70, now.Add(5*time.Second)); got != fdWarnEventWarn {
		t.Errorf("re-crossing threshold after clear: got %v, want warn", got)
	}
}

func TestFdWarnTrackerRewarnInterval(t *testing.T) {
	tracker := newFdWarnTracker(70)
	start := time.Now()

	if got := tracker.observe("sb", 80, start); got != fdWarnEventWarn {
		t.Fatalf("initial cross: got %v, want warn", got)
	}

	// Sustained usage above threshold: silent until the re-warn interval elapses.
	for _, elapsed := range []time.Duration{time.Minute, 5 * time.Minute, fdRewarnInterval - time.Second} {
		if got := tracker.observe("sb", 80, start.Add(elapsed)); got != fdWarnEventNone {
			t.Errorf("at %v: got %v, want none", elapsed, got)
		}
	}

	if got := tracker.observe("sb", 80, start.Add(fdRewarnInterval)); got != fdWarnEventWarn {
		t.Errorf("after re-warn interval: got %v, want warn", got)
	}

	// The window restarts from the second warning: exactly one warn per window.
	if got := tracker.observe("sb", 80, start.Add(fdRewarnInterval+9*time.Minute)); got != fdWarnEventNone {
		t.Errorf("inside second window: got %v, want none", got)
	}
	if got := tracker.observe("sb", 80, start.Add(2*fdRewarnInterval)); got != fdWarnEventWarn {
		t.Errorf("after second window: got %v, want warn", got)
	}
}

func TestFdWarnTrackerPrune(t *testing.T) {
	tracker := newFdWarnTracker(70)
	now := time.Now()

	if got := tracker.observe("gone", 80, now); got != fdWarnEventWarn {
		t.Fatalf("initial cross: got %v, want warn", got)
	}
	if got := tracker.observe("kept", 80, now); got != fdWarnEventWarn {
		t.Fatalf("initial cross: got %v, want warn", got)
	}

	tracker.prune(map[string]struct{}{"kept": {}})

	// A recreated sandbox with the same ID warns fresh.
	if got := tracker.observe("gone", 80, now.Add(time.Second)); got != fdWarnEventWarn {
		t.Errorf("pruned sandbox: got %v, want warn", got)
	}
	// A surviving sandbox keeps its rate-limit state.
	if got := tracker.observe("kept", 80, now.Add(time.Second)); got != fdWarnEventNone {
		t.Errorf("kept sandbox inside re-warn window: got %v, want none", got)
	}
}

func TestSampleRuntimeHelperFds(t *testing.T) {
	root := t.TempDir()
	procDir := filepath.Join(root, "proc")

	// Container init (pid 100) is a child of the runtime helper (pid 50) and
	// a cgroup member: its fds must not be attributed to the helper.
	writeFile(t, filepath.Join(procDir, "100", "stat"), "100 (init) S 50"+strings.Repeat(" 0", 40)+"\n")
	for i := range 9 {
		writeFile(t, filepath.Join(procDir, "100", "fd", strconv.Itoa(i)), "")
	}

	// Helper (pid 50) with 7 fds; its children are the init and a sidecar.
	for i := range 7 {
		writeFile(t, filepath.Join(procDir, "50", "fd", strconv.Itoa(i)), "")
	}
	writeFile(t, filepath.Join(procDir, "50", "task", "50", "children"), "100 51 ")

	// Sidecar helper (pid 51) with 4 fds and no children file.
	for i := range 4 {
		writeFile(t, filepath.Join(procDir, "51", "fd", strconv.Itoa(i)), "")
	}

	total, err := sampleRuntimeHelperFds(procDir, 100, map[int]struct{}{100: {}})
	if err != nil {
		t.Fatalf("sampleRuntimeHelperFds: %v", err)
	}
	if total != 11 {
		t.Errorf("total = %d, want 11 (7 helper + 4 sidecar, cgroup member excluded)", total)
	}
}

func TestSampleRuntimeHelperFdsNoParent(t *testing.T) {
	root := t.TempDir()
	procDir := filepath.Join(root, "proc")

	// Init reparented to pid 1: no helper to attribute to.
	writeFile(t, filepath.Join(procDir, "100", "stat"), "100 (init) S 1"+strings.Repeat(" 0", 40)+"\n")

	if _, err := sampleRuntimeHelperFds(procDir, 100, nil); err == nil {
		t.Error("expected error when init has no runtime helper parent")
	}
}
