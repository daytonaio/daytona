// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package organization

import (
	"bytes"
	"strings"
	"testing"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestRenderTSVInfoEmitsCuratedKeys(t *testing.T) {
	o := &apiclient.Organization{
		Id:        "org-abc",
		Name:      "acme",
		CreatedAt: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	renderTSVInfo(&buf, o)
	out := buf.String()

	pairs := map[string]string{}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		k, v, _ := strings.Cut(line, "\t")
		pairs[k] = v
	}

	want := map[string]string{
		"organization": "acme",
		"created":      "2026-05-01T10:00:00Z",
		"id":           "org-abc",
	}
	for k, v := range want {
		if pairs[k] != v {
			t.Errorf("key %q = %q, want %q (full output:\n%s)", k, pairs[k], v, out)
		}
	}
}
