// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/docker/docker/api/types/container"
)

// GpuIndexLabel is set on every GPU sandbox container with the index of the
// physical GPU the container has been pinned to. The allocator scans this
// label on existing containers to determine which indices are still free.
const GpuIndexLabel = "daytona.gpu_index"

// gpuAllocator hands out GPU device indices to GPU sandboxes on a runner.
// Allocation is serialized by a mutex so concurrent sandbox creations cannot
// pick the same physical card.
type gpuAllocator struct {
	mu    sync.Mutex
	total int
}

func newGpuAllocator(total int) *gpuAllocator {
	return &gpuAllocator{total: total}
}

// Acquire locks the allocator, scans all containers on the runner for the
// daytona.gpu_index label, and returns the lowest free GPU index in
// [0, total). The caller MUST defer the returned release() and MUST call
// ContainerCreate (which sets the label on the new container) BEFORE
// release() runs so concurrent allocators see the new label on their next
// scan.
func (a *gpuAllocator) Acquire(ctx context.Context, d *DockerClient) (int, func(), error) {
	a.mu.Lock()
	release := func() { a.mu.Unlock() }

	if a.total <= 0 {
		release()
		return 0, nil, fmt.Errorf("runner has no GPUs to assign")
	}

	containers, err := d.apiClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		release()
		return 0, nil, fmt.Errorf("list containers for GPU allocation: %w", err)
	}

	// Only containers whose process is alive can actually hold a GPU - Docker
	// detaches the CDI device cgroup on exit, so an exited / dead / removing
	// sandbox no longer occupies its physical card and its index must be
	// reusable by the next allocation. (A subsequent restart of a stopped GPU
	// sandbox is handled at start time by the per-card collision check rather
	// than by keeping the slot reserved here.)
	used := make(map[int]struct{}, len(containers))
	for _, c := range containers {
		switch c.State {
		case "exited", "dead", "removing":
			continue
		}
		if v, ok := c.Labels[GpuIndexLabel]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				used[n] = struct{}{}
			}
		}
	}

	for i := 0; i < a.total; i++ {
		if _, taken := used[i]; !taken {
			return i, release, nil
		}
	}

	release()
	return 0, nil, fmt.Errorf("no free GPU on runner (capacity %d)", a.total)
}
