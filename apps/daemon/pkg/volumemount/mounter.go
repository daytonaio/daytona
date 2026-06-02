// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package volumemount runs inside the sandbox container and mounts layered
// volumes from the env payload injected by the runner. It is the sandbox-side
// counterpart of pkg/volume/incontainer in the runner.
//
// Each volume is mounted via the layered mount CLI as
// `<binary> mount <DISK> <MOUNTPOINT> --region <REGION>`, authenticated by a
// per-(sandbox, volume) token passed via the child env (not argv, so it can't
// leak via /proc/<pid>/cmdline or `ps`).
//
// Env contract (must match pkg/volume/incontainer in runner):
//
//	DAYTONA_INCONTAINER_VOLUMES          JSON-encoded []Volume (per-volume
//	                                     layeredDisk / layeredRegion / layeredMountToken)
//	DAYTONA_INCONTAINER_LAYERED_BINARY   absolute path to the layered mount CLI
//
// The runner provides the privileged FUSE primitives (binary bind-mounted RO,
// /dev/fuse via --device, CAP_SYS_ADMIN), but the snapshot must supply:
//
//   - `fuse3` (the `fusermount3` setuid helper). Some libfuse builds invoke it
//     even as root; missing it fails with "fusermount: executable file not
//     found". preflightCheck warns when it's absent and MountAll appends an
//     install hint to fuse-helper-related failures.
//   - `ca-certificates` for TLS to the layered control plane.
//   - A glibc-compatible runtime (the binary is glibc-linked; Alpine without
//     glibc compat fails to exec it).
package volumemount

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	envVolumesJSON   = "DAYTONA_INCONTAINER_VOLUMES"
	envLayeredBinary = "DAYTONA_INCONTAINER_LAYERED_BINARY"
)

// Volume mirrors volume.Volume in the runner. Only the fields the daemon
// actually consumes are declared here.
type Volume struct {
	VolumeID          string `json:"volumeId"`
	MountPath         string `json:"mountPath"`
	Subpath           string `json:"subpath,omitempty"`
	ReadOnly          bool   `json:"readOnly,omitempty"`
	LayeredDisk       string `json:"layeredDisk,omitempty"`
	LayeredRegion     string `json:"layeredRegion,omitempty"`
	LayeredMountToken string `json:"layeredMountToken,omitempty"`
}

// MountAll reads the env payload and mounts every declared volume,
// idempotently (already-mounted paths are skipped). Each volume is retried up
// to mountMaxAttempts times; if any still fails, MountAll returns an error so
// the daemon exits non-zero rather than coming up with empty mount paths.
//
// The token-bearing env vars are scrubbed before returning so they don't leak
// into child processes spawned later by the daemon or user code.
func MountAll(ctx context.Context, logger *slog.Logger) error {
	defer scrubEnv(logger)

	raw := os.Getenv(envVolumesJSON)
	if raw == "" {
		return nil
	}

	binary := os.Getenv(envLayeredBinary)
	if binary == "" {
		return fmt.Errorf("in-container volume spec present but %s is empty", envLayeredBinary)
	}
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("in-container layered mount binary not found at %q: %w", binary, err)
	}

	var volumes []Volume
	if err := json.Unmarshal([]byte(raw), &volumes); err != nil {
		return fmt.Errorf("parse in-container volume spec: %w", err)
	}
	if len(volumes) == 0 {
		return nil
	}

	if err := preflightCheck(logger); err != nil {
		return err
	}
	helperPresent := hasFuserHelper()

	var failures []error
	for _, v := range volumes {
		if err := mountOneWithRetry(ctx, logger, binary, v); err != nil {
			if !helperPresent && looksLikeFuseHelperError(err) {
				err = fmt.Errorf(
					"%w (hint: fusermount3/fusermount was not found in PATH at startup, "+
						"and the failure looks fuse-helper-related; install the 'fuse3' "+
						"package in your snapshot: apt-get install fuse3 / apk add fuse3 / "+
						"dnf install fuse3)",
					err,
				)
			}
			logger.Error(
				"failed to mount in-container volume after retries",
				"volumeId", v.VolumeID,
				"mountPath", v.MountPath,
				"layeredDisk", v.LayeredDisk,
				"layeredRegion", v.LayeredRegion,
				"attempts", mountMaxAttempts,
				"error", err,
			)
			failures = append(failures, fmt.Errorf("volume %q at %q: %w", v.VolumeID, v.MountPath, err))
		}
	}

	if len(failures) > 0 {
		return errors.Join(failures...)
	}
	return nil
}

const (
	// mountMaxAttempts is the total attempts per volume. Most failures are
	// deterministic (bad token, deleted disk, wrong region), but a short
	// retry window absorbs transient network glitches.
	mountMaxAttempts = 3
	// mountRetryBackoff is the fixed sleep between retries. Exponential
	// backoff is unnecessary: the per-attempt readiness timeout paces us and
	// the 30s mount budget in daemon main caps the total wait.
	mountRetryBackoff = 1 * time.Second
)

// mountOneWithRetry calls mountOne up to mountMaxAttempts times, best-effort
// cleaning up half-mounted state between attempts so a stale mountpoint can't
// fool the next pass.
func mountOneWithRetry(ctx context.Context, logger *slog.Logger, binary string, v Volume) error {
	var lastErr error
	for attempt := 1; attempt <= mountMaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("aborting volume mount on attempt %d/%d: %w", attempt, mountMaxAttempts, err)
		}

		err := mountOne(ctx, logger, binary, v)
		if err == nil {
			if attempt > 1 {
				logger.Info(
					"in-container volume mounted on retry",
					"volumeId", v.VolumeID,
					"mountPath", v.MountPath,
					"attempt", attempt,
				)
			}
			return nil
		}
		lastErr = err

		if attempt == mountMaxAttempts {
			break
		}

		logger.Warn(
			"in-container volume mount failed; will retry",
			"volumeId", v.VolumeID,
			"mountPath", v.MountPath,
			"attempt", attempt,
			"maxAttempts", mountMaxAttempts,
			"error", err,
		)

		// A failed mount may leave the mountpoint half-registered; unmount
		// before retrying so mountOne's "already mounted" shortcut doesn't
		// return a stale success.
		bestEffortUnmount(ctx, logger, binary, v.MountPath)

		select {
		case <-ctx.Done():
			return fmt.Errorf("aborting volume mount during retry backoff: %w (last error: %v)", ctx.Err(), lastErr)
		case <-time.After(mountRetryBackoff):
		}
	}
	return lastErr
}

// bestEffortUnmount tries the CLI's `unmount` subcommand first (clean FUSE
// teardown), then falls back to `umount -l`. Failures are logged and ignored:
// a clean starting state for the retry is preferred but not required.
func bestEffortUnmount(ctx context.Context, logger *slog.Logger, binary string, mountPath string) {
	if !isMountpoint(mountPath) {
		return
	}
	logger.Debug("unmounting half-mounted path before retry", "mountPath", mountPath)

	if err := exec.CommandContext(ctx, binary, "unmount", mountPath).Run(); err == nil {
		return
	}
	if err := exec.CommandContext(ctx, "umount", "-l", mountPath).Run(); err != nil {
		logger.Warn("best-effort unmount failed before retry", "mountPath", mountPath, "error", err)
	}
}

func mountOne(ctx context.Context, logger *slog.Logger, binary string, v Volume) error {
	if v.MountPath == "" {
		return fmt.Errorf("invalid volume entry: empty mountPath")
	}
	if v.LayeredDisk == "" {
		return fmt.Errorf("invalid volume entry: empty layeredDisk")
	}
	if v.LayeredRegion == "" {
		return fmt.Errorf("invalid volume entry: empty layeredRegion")
	}
	if v.LayeredMountToken == "" {
		return fmt.Errorf("invalid volume entry: empty layeredMountToken")
	}

	if err := os.MkdirAll(v.MountPath, 0755); err != nil {
		return fmt.Errorf("create mountpoint: %w", err)
	}

	if isMountpoint(v.MountPath) {
		logger.Debug("volume already mounted in-container", "volumeId", v.VolumeID, "mountPath", v.MountPath)
		return nil
	}

	// The mount CLI's `disk[:/subpath]` syntax mounts a subdirectory of the
	// disk as the root, mirroring NFS conventions.
	target := v.LayeredDisk
	if v.Subpath != "" {
		sub := v.Subpath
		if sub[0] != '/' {
			sub = "/" + sub
		}
		target = v.LayeredDisk + ":" + sub
	}

	args := []string{
		"mount",
		target,
		v.MountPath,
		"--region", v.LayeredRegion,
	}
	if v.ReadOnly {
		args = append(args, "--shared", "--read-only")
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	// Pass the token via env, not argv (argv is visible in
	// /proc/<pid>/cmdline and `ps`). `ARCHIL_MOUNT_TOKEN` is the third-party
	// CLI's own contract (the binary is the `archil` CLI) and the only place
	// we use that name; elsewhere the value travels as `layeredMountToken`.
	cmd.Env = append(os.Environ(), "ARCHIL_MOUNT_TOKEN="+v.LayeredMountToken)

	logger.Info(
		"mounting in-container volume",
		"volumeId", v.VolumeID,
		"mountPath", v.MountPath,
		"layeredDisk", v.LayeredDisk,
		"layeredRegion", v.LayeredRegion,
		"subpath", v.Subpath,
		"readOnly", v.ReadOnly,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("layered mount failed: %w: %s", err, string(out))
	}

	if err := waitUntilReady(ctx, v.MountPath); err != nil {
		return fmt.Errorf("mount not ready: %w", err)
	}
	logger.Info("mounted in-container volume", "volumeId", v.VolumeID, "mountPath", v.MountPath)
	return nil
}

// scrubEnv unsets the token-bearing volume-spec env vars from the daemon's
// own environment so they don't leak into later child processes or env dumps.
// The mounts are unaffected — the forked FUSE server no longer needs the
// token once the CLI returns.
func scrubEnv(logger *slog.Logger) {
	for _, k := range []string{envVolumesJSON, envLayeredBinary} {
		if err := os.Unsetenv(k); err != nil {
			logger.Warn("failed to unset in-container env var", "var", k, "error", err)
		}
	}
}

// preflightCheck validates the OS-level prerequisites for the layered mount:
//
//   - Hard: `/dev/fuse` must exist. The runner attaches it via `--device`, so
//     its absence usually means a runner misconfiguration; we return a
//     distinct error for support to route.
//   - Soft: `fusermount3`/`fusermount`, the libfuse setuid helper. The CLI
//     running as root with CAP_SYS_ADMIN may skip it, so we only warn; if a
//     mount later fails, MountAll uses helperPresent to decide whether to add
//     an "install fuse3" hint.
//
// We deliberately don't check the CA bundle or glibc compat — the CLI's own
// errors (and mountOne's ENOENT wrapping) are clear enough.
func preflightCheck(logger *slog.Logger) error {
	if _, err := os.Stat("/dev/fuse"); err != nil {
		return fmt.Errorf(
			"/dev/fuse is missing inside the sandbox; the kernel FUSE device is "+
				"required for in-container volume mounts. This usually indicates "+
				"the runner did not attach /dev/fuse to the container - please "+
				"contact support: %w", err,
		)
	}

	if !hasFuserHelper() {
		logger.Warn(
			"fusermount3/fusermount not found in $PATH; some libfuse builds " +
				"require it. If volume mounts fail with a 'fusermount: executable " +
				"file not found' or similar message, install the 'fuse3' package " +
				"in your snapshot (apt-get install fuse3 / apk add fuse3 / " +
				"dnf install fuse3).",
		)
	}

	return nil
}

// hasFuserHelper reports whether fusermount3 (fuse3) or the older fusermount
// (fuse 2.x) is reachable via $PATH; libfuse falls back to whichever exists.
func hasFuserHelper() bool {
	for _, name := range []string{"fusermount3", "fusermount"} {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	return false
}

// looksLikeFuseHelperError reports whether the error looks like a missing
// fusermount helper (vs. a bad token, deleted disk, or network error), so we
// only add the "install fuse3" hint when relevant. It matches on text because
// the failure surface is the CLI's combined stdout+stderr.
func looksLikeFuseHelperError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{
		"fusermount",
		"fuse: device not found",
		"fuse_kern_mount",
		"executable file not found",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func isMountpoint(path string) bool {
	cleaned := filepath.Clean(path)
	parent := filepath.Dir(cleaned)

	pi, err := os.Stat(cleaned)
	if err != nil {
		return false
	}
	pp, err := os.Stat(parent)
	if err != nil {
		return false
	}

	pDev, ok1 := statDev(pi)
	parentDev, ok2 := statDev(pp)
	if !ok1 || !ok2 {
		return false
	}
	return pDev != parentDev
}

func waitUntilReady(ctx context.Context, path string) error {
	const maxAttempts = 50
	const sleep = 100 * time.Millisecond

	for i := 0; i < maxAttempts; i++ {
		if !isMountpoint(path) {
			return fmt.Errorf("mount disappeared during readiness check")
		}
		if _, err := os.ReadDir(path); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}
	}
	return fmt.Errorf("mount did not become ready within timeout")
}
