// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"bytes"
	"strings"
	"testing"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestRenderTSVInfoEmitsCuratedKeys(t *testing.T) {
	size := float32(12.5)
	sn := &apiclient.SnapshotDto{
		Id:        "snap-abc",
		Name:      "node-20",
		State:     apiclient.SNAPSHOTSTATE_ACTIVE,
		Size:      *apiclient.NewNullableFloat32(&size),
		CreatedAt: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	renderTSVInfo(&buf, sn)
	out := buf.String()

	pairs := map[string]string{}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		k, v, _ := strings.Cut(line, "\t")
		pairs[k] = v
	}

	want := map[string]string{
		"snapshot": "node-20",
		"state":    "active",
		"size_gb":  "12.50",
		"created":  "2026-05-01T10:00:00Z",
		"id":       "snap-abc",
	}
	for k, v := range want {
		if pairs[k] != v {
			t.Errorf("key %q = %q, want %q (full output:\n%s)", k, pairs[k], v, out)
		}
	}
}

func TestRenderTSVInfoOmitsNilSize(t *testing.T) {
	sn := &apiclient.SnapshotDto{
		Id:        "snap-min",
		Name:      "minimal",
		State:     apiclient.SNAPSHOTSTATE_PENDING,
		CreatedAt: time.Now(),
	}

	var buf bytes.Buffer
	renderTSVInfo(&buf, sn)
	if strings.Contains(buf.String(), "size_gb\t") {
		t.Errorf("nil size should be omitted; got:\n%s", buf.String())
	}
}
