// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) Stop(ctx context.Context, containerId string) error {
	// Deduce sandbox state first
	state, err := d.DeduceSandboxState(ctx, containerId)
	if err == nil && state == enums.SandboxStateStopped {
		log.Debugf("Sandbox %s is already stopped", containerId)
		d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)
		return nil
	}

	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopping)

	if err != nil {
		log.Warnf("Failed to deduce sandbox %s state: %v", containerId, err)
		log.Warnf("Continuing with stop operation")
	}

	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	timeout := 2 // seconds

	// CRITICAL: Copy files from container BEFORE stopping
	// Docker's bind mount might not flush files to host, so we explicitly copy them
	diskId, _ := d.getSandboxDiskId(ctx, containerId)

	// Debug log file
	debugLog, _ := os.OpenFile("/tmp/docker-cp-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if debugLog != nil {
		fmt.Fprintf(debugLog, "\n\n=== STOP CONTAINER %s (disk %s) ===\n", containerId, diskId)
	}

	if diskId != "" {
		disk, err := d.sdisk.Open(ctx, diskId)
		if err == nil {
			// Note: We don't defer disk.Close() here because we might reassign disk later
			// We'll close it manually at the end

			if debugLog != nil {
				fmt.Fprintf(debugLog, "[OPEN] Opened disk, IsMounted=%v\n", disk.IsMounted())
			}

			// Ensure disk is mounted
			if !disk.IsMounted() {
				log.Debugf("Disk not mounted, mounting now...")
				if debugLog != nil {
					fmt.Fprintf(debugLog, "[MOUNT] Disk not mounted, mounting now...\n")
				}
				if _, mErr := disk.Mount(ctx); mErr != nil {
					if debugLog != nil {
						fmt.Fprintf(debugLog, "[MOUNT] Failed to mount: %v\n", mErr)
					}
				}
			}

			// Get mount path for docker cp
			// The disk was mounted during container creation, so the files should already be there
			// We just need the mount path to sync and ensure data is flushed
			mountPath := disk.MountPath()
			if mountPath == "" {
				log.Warnf("Disk %s has no mount path, skipping file copy", diskId)
				if debugLog != nil {
					fmt.Fprintf(debugLog, "[SKIP] No mount path, skipping copy\n")
				}
			} else {
				if debugLog != nil {
					fmt.Fprintf(debugLog, "[MOUNTPATH] Using mountPath=%s\n", mountPath)
				}

				// CRITICAL: Files should already be in the NBD device from the container's bind mount
				// We don't need docker cp - the bind mount already wrote to the NBD device
				// Just verify files are there and sync them
				if debugLog != nil {
					fmt.Fprintf(debugLog, "[CHECK] Checking files in mountPath before sync\n")
				}
				if entries, readErr := os.ReadDir(mountPath); readErr == nil {
					if debugLog != nil {
						fmt.Fprintf(debugLog, "[CHECK] Files in mountPath: %d\n", len(entries))
						for _, entry := range entries {
							info, _ := entry.Info()
							fmt.Fprintf(debugLog, "  - %s (size: %d)\n", entry.Name(), info.Size())
						}
					}
				} else {
					if debugLog != nil {
						fmt.Fprintf(debugLog, "[CHECK] Failed to read mountPath: %v\n", readErr)
					}
				}

				// Use docker cp to copy files from the RUNNING container
				// This handles the case where files were written to the container's overlay
				// instead of the bind mount (e.g., if mount was stale)
				log.Debugf("Copying files from running container %s:/workspace to %s", containerId, mountPath)
				if debugLog != nil {
					fmt.Fprintf(debugLog, "[CP] Copying from %s:/workspace to %s\n", containerId, mountPath)
				}

				cpCmd := exec.CommandContext(ctx, "docker", "cp", containerId+":/workspace/.", mountPath+"/")
				output, cpErr := cpCmd.CombinedOutput()

				if cpErr != nil {
					log.Warnf("Failed to copy files from running container: %v", cpErr)
					if debugLog != nil {
						fmt.Fprintf(debugLog, "[CP] FAILED: %v\n", cpErr)
						fmt.Fprintf(debugLog, "[CP] Output: %s\n", string(output))
					}
				} else {
					log.Debugf("Successfully copied files from running container")
					if debugLog != nil {
						fmt.Fprintf(debugLog, "[CP] SUCCESS\n")
						fmt.Fprintf(debugLog, "[CP] Output: %s\n", string(output))
					}

					// List files after copy
					if entries, readErr := os.ReadDir(mountPath); readErr == nil {
						if debugLog != nil {
							fmt.Fprintf(debugLog, "[CP] Files after copy: %d\n", len(entries))
							for _, entry := range entries {
								info, _ := entry.Info()
								fmt.Fprintf(debugLog, "  - %s (size: %d)\n", entry.Name(), info.Size())
							}
						}
					}

					// Sync immediately
					if err := disk.Sync(ctx); err != nil {
						log.Warnf("Failed to sync after copy: %v", err)
						if debugLog != nil {
							fmt.Fprintf(debugLog, "[SYNC] Failed: %v\n", err)
						}
					} else {
						log.Debugf("Synced filesystem after copy")
						if debugLog != nil {
							fmt.Fprintf(debugLog, "[SYNC] SUCCESS\n")
						}
					}
				}

				// DON'T close the disk here - it will be closed/unmounted later after the container stops
				// Closing here would disconnect the NBD device while the container is still running
				if debugLog != nil {
					fmt.Fprintf(debugLog, "[DONE] Keeping disk open until after container stops\n")
				}
			}
		} else {
			if debugLog != nil {
				fmt.Fprintf(debugLog, "[ERROR] Failed to open disk: %v\n", err)
			}
		}
	}

	if debugLog != nil {
		fmt.Fprintf(debugLog, "[DONE] Pre-stop copy complete\n")
		debugLog.Close()
	}

	// Use exponential backoff helper for container stopping
	err = d.retryWithExponentialBackoff(
		ctx,
		"stop",
		containerId,
		constants.DEFAULT_MAX_RETRIES,
		constants.DEFAULT_BASE_DELAY,
		constants.DEFAULT_MAX_DELAY,
		func() error {
			return d.apiClient.ContainerStop(ctx, containerId, container.StopOptions{
				Signal:  "SIGKILL",
				Timeout: &timeout,
			})
		},
	)
	if err != nil {
		log.Warnf("Failed to stop sandbox %s for %d attempts: %v", containerId, constants.DEFAULT_MAX_RETRIES, err)
		log.Warnf("Trying to kill sandbox %s", containerId)
		err = d.apiClient.ContainerKill(ctx, containerId, "KILL")
		if err != nil {
			log.Warnf("Failed to kill sandbox %s: %v", containerId, err)
		}
		return err
	}

	// Wait for container to actually stop
	statusCh, errCh := d.apiClient.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for sandbox %s to stop: %w", containerId, err)
		}
	case <-statusCh:
		// Container stopped successfully
		d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)

		// IMPORTANT: Don't unmount the disk here - leave it in the pool for reuse
		// The disk will be automatically evicted from the pool when:
		// 1. Another disk needs to be mounted and the pool is full (LRU eviction)
		// 2. The disk is forked (Fork calls pool.Evict for exclusive access)
		// 3. The disk is explicitly deleted
		// This avoids double-unmount issues and allows disk pooling to work efficiently

		log.Debugf("Sandbox %s stopped, disk remains mounted in pool for reuse", containerId)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}

	log.Debugf("Sandbox %s stopped successfully", containerId)
	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)

	// IMPORTANT: Don't unmount the disk here - leave it in the pool for reuse
	// The disk will be automatically evicted from the pool when:
	// 1. Another disk needs to be mounted and the pool is full (LRU eviction)
	// 2. The disk is forked (Fork calls pool.Evict for exclusive access)
	// 3. The disk is explicitly deleted
	// This avoids double-unmount issues and allows disk pooling to work efficiently

	log.Debugf("Sandbox %s stopped, disk remains mounted in pool for reuse", containerId)

	return nil
}
