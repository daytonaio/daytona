// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package metrics

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/prometheus/procfs"
)

const (
	// procRoot and cgroupRoot are the host mount points of procfs and cgroupfs.
	procRoot   = "/proc"
	cgroupRoot = "/sys/fs/cgroup"
)

// fdLimitUnlimited is the value procfs reports for an "unlimited" soft RLIMIT_NOFILE.
const fdLimitUnlimited uint64 = math.MaxUint64

// SandboxFdUsage aggregates host-side file descriptor usage across all
// processes inside a sandbox container's cgroup.
type SandboxFdUsage struct {
	// OpenFds is the total number of open file descriptors summed across all
	// cgroup member processes.
	OpenFds uint64
	// WorstPid is the PID of the member process closest to its soft RLIMIT_NOFILE.
	WorstPid int
	// WorstOpenFds is the number of open file descriptors of the worst process.
	WorstOpenFds uint64
	// WorstFdLimit is the soft RLIMIT_NOFILE of the worst process.
	WorstFdLimit uint64
	// UsagePercent is the highest open/limit ratio among member processes, in
	// percent. Processes with an unlimited RLIMIT_NOFILE contribute to OpenFds
	// but not here, since they cannot hit EMFILE.
	UsagePercent float64
}

// sampleSandboxFdUsage samples file descriptor usage of every process in the
// cgroup of the container whose init process is initPid: the init's cgroup is
// resolved from /proc/<pid>/cgroup, member PIDs are read from cgroup.procs,
// and each member contributes its open fd count and soft RLIMIT_NOFILE.
// Member PIDs that disappear between the cgroup.procs read and per-process
// sampling are skipped (benign race).
func sampleSandboxFdUsage(procfsRoot, cgroupfsRoot string, initPid int) (*SandboxFdUsage, error) {
	fs, err := procfs.NewFS(procfsRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs at %s: %w", procfsRoot, err)
	}

	initProc, err := fs.Proc(initPid)
	if err != nil {
		return nil, fmt.Errorf("failed to access init process %d: %w", initPid, err)
	}

	cgroups, err := initProc.Cgroups()
	if err != nil {
		return nil, fmt.Errorf("failed to read cgroups of pid %d: %w", initPid, err)
	}

	cgroupDir, err := resolveCgroupDir(cgroups, cgroupfsRoot)
	if err != nil {
		return nil, err
	}

	pids, err := readCgroupProcs(filepath.Join(cgroupDir, "cgroup.procs"))
	if err != nil {
		return nil, err
	}

	usage := &SandboxFdUsage{}
	for _, pid := range pids {
		proc, err := fs.Proc(pid)
		if err != nil {
			continue
		}

		openFds, err := countOpenFds(procfsRoot, proc)
		if err != nil {
			continue
		}

		limits, err := proc.Limits()
		if err != nil {
			continue
		}

		usage.OpenFds += openFds

		if limits.OpenFiles == 0 || limits.OpenFiles == fdLimitUnlimited {
			continue
		}

		percent := float64(openFds) / float64(limits.OpenFiles) * 100
		if usage.WorstPid == 0 || percent > usage.UsagePercent {
			usage.WorstPid = pid
			usage.WorstOpenFds = openFds
			usage.WorstFdLimit = limits.OpenFiles
			usage.UsagePercent = percent
		}
	}

	return usage, nil
}

// countOpenFds returns the number of open file descriptors of proc. On Linux
// it delegates to procfs, which stats /proc/<pid>/fd (the kernel reports the
// entry count as the directory size since v6.2) and otherwise falls back to a
// single Readdirnames pass. On other platforms (unit tests over fake /proc
// trees) procfs mistakes the fake tree for a real procfs and would return the
// directory byte size, so the entries are counted directly.
func countOpenFds(procfsRoot string, proc procfs.Proc) (uint64, error) {
	if runtime.GOOS == "linux" {
		openFds, err := proc.FileDescriptorsLen()
		if err != nil {
			return 0, err
		}
		return uint64(openFds), nil
	}

	dir, err := os.Open(filepath.Join(procfsRoot, strconv.Itoa(proc.PID), "fd"))
	if err != nil {
		return 0, err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return 0, err
	}

	return uint64(len(names)), nil
}

// resolveCgroupDir maps /proc/<pid>/cgroup entries to the cgroupfs directory
// holding the process. The cgroup v2 unified hierarchy (HierarchyID 0) is
// preferred; on cgroup v1 the hierarchy holding the pids controller is used.
func resolveCgroupDir(cgroups []procfs.Cgroup, cgroupfsRoot string) (string, error) {
	for _, cg := range cgroups {
		if cg.HierarchyID == 0 {
			return filepath.Join(cgroupfsRoot, cg.Path), nil
		}
	}

	for _, cg := range cgroups {
		for _, controller := range cg.Controllers {
			if controller == "pids" {
				return filepath.Join(cgroupfsRoot, "pids", cg.Path), nil
			}
		}
	}

	return "", errors.New("no unified or pids cgroup hierarchy found")
}

// readCgroupProcs parses a cgroup.procs file into its member PIDs.
func readCgroupProcs(path string) ([]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pids []int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		pid, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pid %q in %s: %w", line, path, err)
		}

		pids = append(pids, pid)
	}

	return pids, scanner.Err()
}
