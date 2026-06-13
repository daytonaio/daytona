// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestDeleteIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "plain error", err: errors.New("boom"), want: false},
		{name: "not found clierr", err: clierr.New(clierr.CategoryNotFound, "snapshot not found"), want: true},
		{name: "other category clierr", err: clierr.New(clierr.CategoryConflict, "in use"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deleteIsNotFound(tt.err); got != tt.want {
				t.Errorf("deleteIsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestSnapshotDryRunResult(t *testing.T) {
	items := []apiclient.SnapshotDto{
		{Id: "snap-1", Name: "base", State: apiclient.SNAPSHOTSTATE_ACTIVE},
		{Id: "snap-2", Name: "stale", State: apiclient.SNAPSHOTSTATE_INACTIVE},
	}

	result := snapshotDryRunResult(items)

	if !result.DryRun {
		t.Error("DryRun = false, want true")
	}
	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}
	if len(result.Snapshots) != 2 {
		t.Fatalf("len(Snapshots) = %d, want 2", len(result.Snapshots))
	}
	if result.Snapshots[0].Id != "snap-1" || result.Snapshots[0].Name != "base" || result.Snapshots[0].State != "active" {
		t.Errorf("Snapshots[0] = %+v, want {snap-1 base active}", result.Snapshots[0])
	}
}

func TestSnapshotDryRunResultUsesSnapshotsKey(t *testing.T) {
	data, err := json.Marshal(snapshotDryRunResult(nil))
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if !strings.Contains(string(data), `"snapshots":[]`) {
		t.Errorf("marshaled dry-run result %s missing empty snapshots array", data)
	}
}

func TestNewDeleteBulkResultJSONShape(t *testing.T) {
	data, err := json.Marshal(newDeleteBulkResult(3))
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	for _, want := range []string{`"dryRun":false`, `"count":3`, `"deleted":[]`, `"failed":[]`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("marshaled bulk result %s missing %s", data, want)
		}
	}
}
