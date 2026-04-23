// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/stretchr/testify/require"
)

func writeJSONResponse(t *testing.T, w http.ResponseWriter, status int, payload any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	require.NoError(t, json.NewEncoder(w).Encode(payload))
}

func testSandboxPayload(id, name string, state apiclient.SandboxState) map[string]any {
	return map[string]any{
		"id":                  id,
		"organizationId":      "org-1",
		"name":                name,
		"user":                "daytona",
		"env":                 map[string]string{"A": "B"},
		"labels":              map[string]string{types.CodeToolboxLanguageLabel: string(types.CodeLanguagePython)},
		"public":              false,
		"networkBlockAll":     false,
		"target":              "us-east-1",
		"cpu":                 1,
		"gpu":                 0,
		"memory":              2,
		"disk":                10,
		"state":               state,
		"toolboxProxyUrl":     "http://toolbox-proxy.test",
		"autoArchiveInterval": 15,
		"autoDeleteInterval":  -1,
	}
}

func testSnapshotPayload(id, name string, state apiclient.SnapshotState) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]any{
		"id":             id,
		"organizationId": "org-1",
		"general":        false,
		"name":           name,
		"imageName":      "python:3.12",
		"state":          state,
		"size":           12.5,
		"entrypoint":     []string{"python", "app.py"},
		"cpu":            2,
		"gpu":            0,
		"mem":            4096,
		"disk":           20,
		"errorReason":    nil,
		"createdAt":      now,
		"updatedAt":      now,
		"lastUsedAt":     nil,
	}
}

func testVolumePayload(id, name string, state apiclient.VolumeState) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]any{
		"id":             id,
		"name":           name,
		"organizationId": "org-1",
		"state":          state,
		"createdAt":      now,
		"updatedAt":      now,
		"errorReason":    nil,
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}

	if s, ok := v.([]string); ok {
		return s
	}

	if ifaceSlice, ok := v.([]any); ok {
		out := make([]string, 0, len(ifaceSlice))
		for _, item := range ifaceSlice {
			out = append(out, fmt.Sprint(item))
		}
		return out
	}

	return []string{fmt.Sprint(v)}
}
