// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"bytes"
	"strings"
	"testing"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestRenderTSVInfoEmitsCuratedKeys(t *testing.T) {
	v := &apiclient.VolumeDto{
		Id:        "vol-abc",
		Name:      "shared-cache",
		State:     apiclient.VOLUMESTATE_READY,
		CreatedAt: "2026-05-01T10:00:00Z",
	}

	var buf bytes.Buffer
	renderTSVInfo(&buf, v)
	out := buf.String()

	pairs := map[string]string{}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		k, val, _ := strings.Cut(line, "\t")
		pairs[k] = val
	}

	want := map[string]string{
		"volume":  "shared-cache",
		"id":      "vol-abc",
		"state":   "ready",
		"created": "2026-05-01T10:00:00Z",
	}
	for k, v := range want {
		if pairs[k] != v {
			t.Errorf("key %q = %q, want %q (full output:\n%s)", k, pairs[k], v, out)
		}
	}
}
