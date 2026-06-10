// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"encoding/json"
	"errors"
	"fmt"
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
		{name: "not found clierr", err: clierr.New(clierr.CategoryNotFound, "sandbox not found"), want: true},
		{name: "other category clierr", err: clierr.New(clierr.CategoryServer, "internal error"), want: false},
		{name: "wrapped not found", err: fmt.Errorf("context: %w", clierr.New(clierr.CategoryNotFound, "gone")), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deleteIsNotFound(tt.err); got != tt.want {
				t.Errorf("deleteIsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestSandboxDryRunResult(t *testing.T) {
	state := apiclient.SANDBOXSTATE_STARTED
	items := []apiclient.SandboxListItem{
		{Id: "id-1", Name: "one", State: &state},
		{Id: "id-2", Name: "two"},
	}

	result := sandboxDryRunResult(items)

	if !result.DryRun {
		t.Error("DryRun = false, want true")
	}
	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}
	if len(result.Sandboxes) != 2 {
		t.Fatalf("len(Sandboxes) = %d, want 2", len(result.Sandboxes))
	}
	if result.Sandboxes[0].Id != "id-1" || result.Sandboxes[0].Name != "one" || result.Sandboxes[0].State != "started" {
		t.Errorf("Sandboxes[0] = %+v, want {id-1 one started}", result.Sandboxes[0])
	}
	if result.Sandboxes[1].State != "" {
		t.Errorf("Sandboxes[1].State = %q, want empty for nil state", result.Sandboxes[1].State)
	}
}

func TestSandboxDryRunResultEmptyMarshalsArray(t *testing.T) {
	data, err := json.Marshal(sandboxDryRunResult(nil))
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if !strings.Contains(string(data), `"sandboxes":[]`) {
		t.Errorf("marshaled dry-run result %s missing empty sandboxes array", data)
	}
}

func TestNewDeleteBulkResultJSONShape(t *testing.T) {
	data, err := json.Marshal(newDeleteBulkResult(0))
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	for _, want := range []string{`"dryRun":false`, `"count":0`, `"deleted":[]`, `"failed":[]`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("marshaled bulk result %s missing %s", data, want)
		}
	}
}

func TestDeleteSingleResultJSONShape(t *testing.T) {
	name := "my-sandbox"
	tests := []struct {
		name   string
		result deleteSingleResult
		want   []string
	}{
		{
			name:   "deleted sandbox",
			result: deleteSingleResult{Id: "id-1", Name: &name, Deleted: true, Found: true},
			want:   []string{`"id":"id-1"`, `"name":"my-sandbox"`, `"deleted":true`, `"found":true`},
		},
		{
			name:   "not found sandbox",
			result: deleteSingleResult{Id: "missing", Name: nil, Deleted: false, Found: false},
			want:   []string{`"id":"missing"`, `"name":null`, `"deleted":false`, `"found":false`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.result)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			for _, want := range tt.want {
				if !strings.Contains(string(data), want) {
					t.Errorf("marshaled result %s missing %s", data, want)
				}
			}
		})
	}
}
