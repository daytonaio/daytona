// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package volumemount runs inside the sandbox container and performs the
// in-container volume mount driven by the env payload injected by the runner.
// This is the sandbox-side counterpart of pkg/volume/incontainer in the
// runner.
//
// Backend: Archil. Each volume is mounted with `archil mount <DISK>
// <MOUNTPOINT> --region <REGION>`, authenticated by a per-disk
// ARCHIL_MOUNT_TOKEN passed via the child process environment (never on the
// command line, so it can't leak via /proc/<pid>/cmdline or `ps`).
//
// Env contract (must match pkg/volume/incontainer in runner):
//
//	DAYTONA_INCONTAINER_VOLUMES         JSON-encoded []Volume (with per-volume
//	                                    archilDisk / archilRegion / archilMountToken)
//	DAYTONA_INCONTAINER_ARCHIL_BINARY   absolute path to the archil CLI binary
//
// Snapshot prerequisites:
//
// The runner provides the privileged primitives needed to mount FUSE inside
// the sandbox (the `archil` binary bind-mounted RO, `/dev/fuse` attached
// via --device, and a privileged container with CAP_SYS_ADMIN), but the
// snapshot must supply the userspace pieces archil's libfuse build talks to:
//
//   - `fuse3` package (provides the `fusermount3` setuid helper). Some
//     libfuse builds invoke fusermount3 even when running as root; missing
//     it causes mounts to fail with "fusermount: executable file not found".
//     preflightCheck warns if neither fusermount3 nor fusermount is in PATH,
//     and MountAll appends an actionable hint to the surfaced error if the
//     subsequent mount failure looks fuse-helper-related.
//   - `ca-certificates` for TLS to the Archil control plane.
//   - A glibc-compatible runtime (the archil binary is dynamically linked
//     against glibc; Alpine snapshots without glibc compat will fail to
//     exec the bind-mounted binary).
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
	envVolumesJSON  = "DAYTONA_INCONTAINER_VOLUMES"
	envArchilBinary = "DAYTONA_INCONTAINER_ARCHIL_BINARY"
)

// Volume mirrors volume.Volume in the runner. Only the fields the daemon
// actually consumes are declared here.
type Volume struct {
	VolumeID         string `json:"volumeId"`
	MountPath        string `json:"mountPath"`
	Subpath          string `json:"subpath,omitempty"`
	ReadOnly         bool   `json:"readOnly,omitempty"`
	ArchilDisk       string `json:"archilDisk,omitempty"`
	ArchilRegion     string `json:"archilRegion,omitempty"`
	ArchilMountToken string `json:"archilMountToken,omitempty"`
}

// MountAll reads the env payload and mounts every declared volume. It is
// idempotent — already-mounted paths are skipped.
//
// Each volume is attempted up to mountMaxAttempts times before giving up.
// If any volume fails after all retries, MountAll returns an error so the
// daemon can exit non-zero and the runner can surface the failure as a
// sandbox-level error rather than letting the sandbox come up with empty
// mount paths.
//
// As a defensive measure the env vars carrying the volume spec (which contain
// per-disk Archil mount tokens) are scrubbed from the daemon's own process
// environment before returning. Child processes spawned later by the daemon
// or by user code will not inherit them.
func MountAll(ctx context.Context, logger *slog.Logger) error {
	defer scrubEnv(logger)

	raw := os.Getenv(envVolumesJSON)
	if raw == "" {
		return nil
	}

	binary := os.Getenv(envArchilBinary)
	if binary == "" {
		return fmt.Errorf("in-container volume spec present but %s is empty", envArchilBinary)
	}
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("in-container archil binary not found at %q: %w", binary, err)
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
						"package in your snapshot: apt-get install fuse3 / apk add fuse3 / dnf install fuse3)",
					err,
				)
			}
			logger.Error(
				"failed to mount in-container volume after retries",
				"volumeId", v.VolumeID,
				"mountPath", v.MountPath,
				"archilDisk", v.ArchilDisk,
				"archilRegion", v.ArchilRegion,
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
	// mountMaxAttempts is the total number of attempts per volume (one
	// initial attempt + retries). Most failures are deterministic
	// (bad token, deleted disk, wrong region) so retries don't help, but a
	// short retry window absorbs transient network glitches without making
	// users wait long when the failure is permanent.
	mountMaxAttempts = 3
	// mountRetryBackoff is the fixed sleep between retry attempts. We don't
	// bother with exponential backoff because the per-attempt 5s readiness
	// timeout already paces us, and the overall mount budget (30s in
	// daemon main) caps the total wait.
	mountRetryBackoff = 1 * time.Second
)

// mountOneWithRetry calls mountOne up to mountMaxAttempts times. Between
// attempts it best-effort-cleans up any half-mounted state so the next
// attempt isn't fooled by a stale mountpoint.
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

		// A failed `archil mount` may have left the FUSE mountpoint
		// half-registered. Attempt a best-effort unmount before retrying
		// so mountOne's "already mounted" shortcut doesn't return a
		// stale success on the next pass.
		bestEffortUnmount(ctx, logger, binary, v.MountPath)

		select {
		case <-ctx.Done():
			return fmt.Errorf("aborting volume mount during retry backoff: %w (last error: %v)", ctx.Err(), lastErr)
		case <-time.After(mountRetryBackoff):
		}
	}
	return lastErr
}

// bestEffortUnmount tries `archil unmount` first (which flushes pending
// writes and tears down the FUSE server cleanly), and falls back to
// `umount -l` if archil didn't manage. Any failure is logged at Warn and
// otherwise ignored — the caller is about to retry the mount, and starting
// from a clean state is preferred but not required.
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
	if v.ArchilDisk == "" {
		return fmt.Errorf("invalid volume entry: empty archilDisk")
	}
	if v.ArchilRegion == "" {
		return fmt.Errorf("invalid volume entry: empty archilRegion")
	}
	if v.ArchilMountToken == "" {
		return fmt.Errorf("invalid volume entry: empty archilMountToken")
	}

	if err := os.MkdirAll(v.MountPath, 0755); err != nil {
		return fmt.Errorf("create mountpoint: %w", err)
	}

	if isMountpoint(v.MountPath) {
		logger.Debug("volume already mounted in-container", "volumeId", v.VolumeID, "mountPath", v.MountPath)
		return nil
	}

	// archil supports `disk[:/subpath]` syntax to mount a subdirectory of
	// the disk as the mount root, mirroring NFS conventions.
	target := v.ArchilDisk
	if v.Subpath != "" {
		sub := v.Subpath
		if sub[0] != '/' {
			sub = "/" + sub
		}
		target = v.ArchilDisk + ":" + sub
	}

	args := []string{
		"mount",
		target,
		v.MountPath,
		"--region", v.ArchilRegion,
	}
	if v.ReadOnly {
		// `--read-only` was added in archil client v0.5.0. Read-only
		// mounts don't take a write delegation, so multiple sandboxes
		// can hold concurrent RO views of the same disk while a separate
		// RW mount is active elsewhere.
		args = append(args, "--read-only")
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	// Pass the token via env, not argv: argv is visible in /proc/<pid>/cmdline
	// and `ps`, env (for processes the daemon doesn't own) is not. The archil
	// CLI itself reads ARCHIL_MOUNT_TOKEN from env.
	cmd.Env = append(os.Environ(), "ARCHIL_MOUNT_TOKEN="+v.ArchilMountToken)

	logger.Info(
		"mounting in-container volume",
		"volumeId", v.VolumeID,
		"mountPath", v.MountPath,
		"archilDisk", v.ArchilDisk,
		"archilRegion", v.ArchilRegion,
		"subpath", v.Subpath,
		"readOnly", v.ReadOnly,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("archil mount failed: %w: %s", err, string(out))
	}

	if err := waitUntilReady(ctx, v.MountPath); err != nil {
		return fmt.Errorf("mount not ready: %w", err)
	}
	logger.Info("mounted in-container volume", "volumeId", v.VolumeID, "mountPath", v.MountPath)
	return nil
}

// scrubEnv unsets the env vars that carry the volume spec (which contain
// per-disk mount tokens) from the daemon's own process environment, so they
// don't leak into child processes spawned later or get printed by anything
// that dumps `os.Environ()`.
//
// The archil mounts themselves are unaffected — once `archil mount` returns,
// the FUSE server it forked off no longer needs ARCHIL_MOUNT_TOKEN.
func scrubEnv(logger *slog.Logger) {
	for _, k := range []string{envVolumesJSON, envArchilBinary} {
		if err := os.Unsetenv(k); err != nil {
			logger.Warn("failed to unset in-container env var", "var", k, "error", err)
		}
	}
}

// preflightCheck validates that the container has the OS-level prerequisites
// for `archil mount` to succeed.
//
// We split prerequisites into two tiers:
//
//   - Hard requirements (mount cannot work without these). Right now this is
//     only `/dev/fuse`. The runner attaches it via `--device`, so a missing
//     device almost always indicates a runner-level misconfiguration rather
//     than a snapshot problem; we surface a distinct error so support can
//     route it correctly.
//
//   - Soft requirements (warned about, not enforced). `fusermount3` /
//     `fusermount` is the userspace setuid helper libfuse uses when it
//     can't (or wasn't built to) call mount(2) directly. archil running as
//     root with CAP_SYS_ADMIN may take the direct path and skip the helper
//     entirely, so we don't want to hard-fail snapshots that work fine
//     without it. If a mount later fails, MountAll uses helperPresent to
//     decide whether to enrich the error with a "install fuse3" hint.
//
// Other less-obvious snapshot dependencies that archil needs at runtime
// but that we deliberately do NOT check here:
//
//   - CA bundle for TLS to the Archil control plane (typically
//     `ca-certificates`). archil's own error output is clear enough.
//   - glibc-compatible libc (the archil binary is glibc-linked). Missing
//     loader / wrong libc surfaces as ENOENT on the binary, which mountOne
//     already wraps clearly.
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

// hasFuserHelper returns true if either fusermount3 (fuse3) or the older
// fusermount (fuse 2.x) helper is reachable via $PATH. Both are acceptable
// because libfuse will fall back to whichever is available.
func hasFuserHelper() bool {
	for _, name := range []string{"fusermount3", "fusermount"} {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	return false
}

// looksLikeFuseHelperError reports whether the error chain looks like a
// failure caused by a missing fusermount helper (rather than e.g. a bad
// token, deleted disk, or network error). We use this together with
// hasFuserHelper to decide whether to append the "install fuse3" hint -
// avoiding misleading hints on unrelated failures like authentication
// errors.
//
// We pattern-match on text rather than typed errors because the failure
// surface is `archil mount`'s combined stdout+stderr, which we capture as
// a wrapped error in mountOne.
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
