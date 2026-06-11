// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package metrics

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

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
