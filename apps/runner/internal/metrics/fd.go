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
	"time"

	"github.com/prometheus/procfs"
)

const (
	// procRoot and cgroupRoot are the host mount points of procfs and cgroupfs.
	procRoot   = "/proc"
	cgroupRoot = "/sys/fs/cgroup"

	// fdWarnClearMarginPercent is the hysteresis band below the warning
	// threshold: a warned sandbox must drop more than this many percentage
	// points below the threshold before its warning state clears.
	fdWarnClearMarginPercent = 5.0
	// fdRewarnInterval is how long a sandbox staying above the threshold is
	// kept silent after a warning before it is warned about again.
	fdRewarnInterval = 10 * time.Minute
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
	// memberPids is the set of PIDs listed in cgroup.procs, used to exclude
	// already-attributed processes from runtime-helper fd attribution.
	memberPids map[int]struct{}
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

	usage := &SandboxFdUsage{memberPids: make(map[int]struct{}, len(pids))}
	for _, pid := range pids {
		usage.memberPids[pid] = struct{}{}

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

// sampleRuntimeHelperFds best-effort sums the open file descriptors of the
// runtime helper processes (containerd-shim, conmon and friends) serving the
// container whose init process is initPid: the helper tree is init's parent
// and that parent's descendants, minus the processes already attributed
// through the container cgroup. Descendants are discovered via
// /proc/<pid>/task/<tid>/children, which requires CONFIG_PROC_CHILDREN.
func sampleRuntimeHelperFds(procfsRoot string, initPid int, cgroupMembers map[int]struct{}) (uint64, error) {
	fs, err := procfs.NewFS(procfsRoot)
	if err != nil {
		return 0, fmt.Errorf("failed to open procfs at %s: %w", procfsRoot, err)
	}

	initProc, err := fs.Proc(initPid)
	if err != nil {
		return 0, fmt.Errorf("failed to access init process %d: %w", initPid, err)
	}

	stat, err := initProc.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to read stat of pid %d: %w", initPid, err)
	}

	if stat.PPID <= 1 {
		return 0, fmt.Errorf("init process %d has no runtime helper parent", initPid)
	}

	var total uint64
	visited := make(map[int]struct{})
	queue := []int{stat.PPID}
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]

		if _, ok := visited[pid]; ok {
			continue
		}
		visited[pid] = struct{}{}

		// Cgroup members (and their descendants, which share the cgroup) are
		// already attributed through sampleSandboxFdUsage.
		if _, member := cgroupMembers[pid]; member {
			continue
		}

		if proc, err := fs.Proc(pid); err == nil {
			if openFds, err := countOpenFds(procfsRoot, proc); err == nil {
				total += openFds
			}
		}

		children, err := readProcChildren(procfsRoot, pid)
		if err != nil {
			continue
		}
		queue = append(queue, children...)
	}

	return total, nil
}

// readProcChildren collects the child PIDs of a process from its
// /proc/<pid>/task/<tid>/children files.
func readProcChildren(procfsRoot string, pid int) ([]int, error) {
	taskDir := filepath.Join(procfsRoot, strconv.Itoa(pid), "task")
	tids, err := os.ReadDir(taskDir)
	if err != nil {
		return nil, err
	}

	var children []int
	for _, tid := range tids {
		data, err := os.ReadFile(filepath.Join(taskDir, tid.Name(), "children"))
		if err != nil {
			continue
		}

		for _, field := range strings.Fields(string(data)) {
			child, err := strconv.Atoi(field)
			if err != nil {
				continue
			}
			children = append(children, child)
		}
	}

	return children, nil
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

// fdWarnEvent is the outcome of observing a sandbox fd usage sample.
type fdWarnEvent int

const (
	fdWarnEventNone fdWarnEvent = iota
	// fdWarnEventWarn fires when usage crosses the threshold, or stays above
	// it for longer than fdRewarnInterval since the last warning.
	fdWarnEventWarn
	// fdWarnEventClear fires once when usage of a warned sandbox drops below
	// the hysteresis band.
	fdWarnEventClear
)

type fdWarnState struct {
	active     bool
	lastWarnAt time.Time
}

// fdWarnTracker rate-limits per-sandbox fd usage warnings: a sandbox is
// warned about when its usage crosses thresholdPercent, re-warned at most
// every fdRewarnInterval while it stays above, and re-armed only after usage
// drops below thresholdPercent-fdWarnClearMarginPercent (hysteresis), so
// usage oscillating around the threshold cannot spam the log.
type fdWarnTracker struct {
	thresholdPercent float64
	states           map[string]*fdWarnState
}

func newFdWarnTracker(thresholdPercent float64) *fdWarnTracker {
	return &fdWarnTracker{
		thresholdPercent: thresholdPercent,
		states:           make(map[string]*fdWarnState),
	}
}

// observe records a usage sample for a sandbox and reports whether the caller
// should emit a warning or a recovery notice. The clock is injected via now.
func (t *fdWarnTracker) observe(sandboxId string, percent float64, now time.Time) fdWarnEvent {
	state := t.states[sandboxId]

	if percent >= t.thresholdPercent {
		if state == nil {
			t.states[sandboxId] = &fdWarnState{active: true, lastWarnAt: now}
			return fdWarnEventWarn
		}
		if !state.active || now.Sub(state.lastWarnAt) >= fdRewarnInterval {
			state.active = true
			state.lastWarnAt = now
			return fdWarnEventWarn
		}
		return fdWarnEventNone
	}

	if state != nil && state.active && percent < t.thresholdPercent-fdWarnClearMarginPercent {
		state.active = false
		return fdWarnEventClear
	}

	return fdWarnEventNone
}

// prune drops state for sandboxes absent from seen so a recreated sandbox
// warns fresh.
func (t *fdWarnTracker) prune(seen map[string]struct{}) {
	for sandboxId := range t.states {
		if _, ok := seen[sandboxId]; !ok {
			delete(t.states, sandboxId)
		}
	}
}
