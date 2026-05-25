// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"os/exec"
	"strings"
)

// detectGpus probes the host for NVIDIA GPUs by invoking
// `nvidia-smi --query-gpu=name --format=csv,noheader` and returns the number
// of GPUs found together with the name of the first one. When nvidia-smi is
// not installed or returns an error (which is the case on CPU-only hosts),
// (0, "") is returned without surfacing the error - the runner simply has no
// GPUs to schedule.
func detectGpus(ctx context.Context) (int, string) {
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	out, err := cmd.Output()
	if err != nil {
		return 0, ""
	}

	var lines []string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return 0, ""
	}
	return len(lines), lines[0]
}
