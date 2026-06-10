// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

// stateTestSandboxJSON renders a minimal Sandbox payload containing every
// property the generated client requires plus the given state.
func stateTestSandboxJSON(state string) string {
	return fmt.Sprintf(`{
		"id": "sbx-1",
		"organizationId": "org-1",
		"name": "my-sandbox",
		"user": "daytona",
		"env": {},
		"labels": {},
		"public": false,
		"networkBlockAll": false,
		"target": "us",
		"cpu": 1,
		"gpu": 0,
		"memory": 1,
		"disk": 1,
		"toolboxProxyUrl": "",
		"state": %q
	}`, state)
}

func stateTestApiClient(serverURL string) *apiclient.APIClient {
	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{{URL: serverURL}}
	return apiclient.NewAPIClient(clientConfig)
}

func stateTestServer(t *testing.T, state string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, stateTestSandboxJSON(state))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestAwaitSandboxStateTimeout(t *testing.T) {
	server := stateTestServer(t, "starting")
	apiClient := stateTestApiClient(server.URL)

	start := time.Now()
	err := common.AwaitSandboxState(context.Background(), apiClient, "sbx-1", 50*time.Millisecond, apiclient.SANDBOXSTATE_STARTED)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	var cliErr *clierr.Error
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *clierr.Error, got %T: %v", err, err)
	}
	if cliErr.Category != clierr.CategoryTimeout {
		t.Errorf("category = %q, want %q", cliErr.Category, clierr.CategoryTimeout)
	}
	if elapsed >= time.Second {
		t.Errorf("timeout took %s, expected to expire well before the next 1s poll", elapsed)
	}
}

func TestAwaitSandboxStateSuccess(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{name: "bounded wait reaches target state", timeout: 5 * time.Second},
		{name: "unbounded wait reaches target state", timeout: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := stateTestServer(t, "started")
			apiClient := stateTestApiClient(server.URL)

			err := common.AwaitSandboxState(context.Background(), apiClient, "sbx-1", tt.timeout, apiclient.SANDBOXSTATE_STARTED)
			if err != nil {
				t.Fatalf("AwaitSandboxState() unexpected error: %v", err)
			}
		})
	}
}

func TestAwaitSandboxStateContextCanceled(t *testing.T) {
	server := stateTestServer(t, "starting")
	apiClient := stateTestApiClient(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := common.AwaitSandboxState(ctx, apiClient, "sbx-1", 0, apiclient.SANDBOXSTATE_STARTED)
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}
}
