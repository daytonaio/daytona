// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSandboxConstruction(t *testing.T) {
	tests := []struct {
		name                string
		id                  string
		sandboxName         string
		state               apiclient.SandboxState
		target              string
		autoArchiveInterval int
		autoDeleteInterval  int
		networkBlockAll     bool
		networkAllowList    *string
	}{
		{
			name:                "basic construction",
			id:                  "test-id",
			sandboxName:         "test-name",
			state:               apiclient.SANDBOXSTATE_STARTED,
			target:              "us-east-1",
			autoArchiveInterval: 60,
			autoDeleteInterval:  -1,
			networkBlockAll:     false,
			networkAllowList:    nil,
		},
		{
			name:                "with network allow list",
			id:                  "id-2",
			sandboxName:         "sandbox-2",
			state:               apiclient.SANDBOXSTATE_STOPPED,
			target:              "eu-west-1",
			autoArchiveInterval: 0,
			autoDeleteInterval:  0,
			networkBlockAll:     true,
			networkAllowList:    strPtr("10.0.0.0/8"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("DAYTONA_API_KEY", "test-api-key")

			client, err := NewClient()
			require.NoError(t, err)

			sandbox := NewSandbox(
				client,
				nil,
				tt.id,
				tt.sandboxName,
				tt.state,
				tt.target,
				tt.autoArchiveInterval,
				tt.autoDeleteInterval,
				tt.networkBlockAll,
				tt.networkAllowList,
			)

			require.NotNil(t, sandbox)
			assert.Equal(t, tt.id, sandbox.ID)
			assert.Equal(t, tt.sandboxName, sandbox.Name)
			assert.Equal(t, tt.state, sandbox.State)
			assert.Equal(t, tt.target, sandbox.Target)
			assert.Equal(t, tt.autoArchiveInterval, sandbox.AutoArchiveInterval)
			assert.Equal(t, tt.autoDeleteInterval, sandbox.AutoDeleteInterval)
			assert.Equal(t, tt.networkBlockAll, sandbox.NetworkBlockAll)
			assert.Equal(t, tt.networkAllowList, sandbox.NetworkAllowList)

			assert.NotNil(t, sandbox.FileSystem)
			assert.NotNil(t, sandbox.Git)
			assert.NotNil(t, sandbox.Process)
			assert.NotNil(t, sandbox.CodeInterpreter)
			assert.NotNil(t, sandbox.ComputerUse)
		})
	}
	os.Clearenv()
}

func TestSandboxStartTimeoutValidation(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STOPPED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err = sandbox.doStartWithTimeout(ctx, -1*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Timeout must be a non-negative number")

	os.Clearenv()
}

func TestSandboxStopTimeoutValidation(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err = sandbox.doStopWithTimeout(ctx, -1*time.Second, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Timeout must be a non-negative number")

	os.Clearenv()
}

func TestSandboxDeleteTimeoutValidation(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STOPPED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err = sandbox.doDeleteWithTimeout(ctx, -1*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Timeout must be a non-negative number")

	os.Clearenv()
}

func TestSandboxWaitForStartTimeoutValidation(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_CREATING, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err = sandbox.doWaitForStart(ctx, -1*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Timeout must be a non-negative number")

	os.Clearenv()
}

func TestSandboxWaitForStopTimeoutValidation(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STOPPING, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err = sandbox.doWaitForStop(ctx, -1*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Timeout must be a non-negative number")

	os.Clearenv()
}

func TestSandboxSetAutoArchiveIntervalNil(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err = sandbox.doSetAutoArchiveInterval(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "intervalMinutes cannot be nil")

	os.Clearenv()
}

func TestSandboxSetAutoDeleteIntervalNil(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err = sandbox.doSetAutoDeleteInterval(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "intervalMinutes cannot be nil")

	os.Clearenv()
}

func TestSandboxResizeNilResources(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err = sandbox.doResizeWithTimeout(ctx, nil, 60*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Resources must not be nil")

	os.Clearenv()
}

func TestSandboxResizeTimeoutValidation(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err = sandbox.doResizeWithTimeout(ctx, &types.Resources{CPU: 2}, -1*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Timeout must be a non-negative number")

	os.Clearenv()
}

func TestSandboxStartAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "start failed"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STOPPED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err := sandbox.doStartWithTimeout(ctx, 5*time.Second)
	require.Error(t, err)

	os.Clearenv()
}

func TestSandboxStopAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "stop failed"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err := sandbox.doStopWithTimeout(ctx, 5*time.Second, false)
	require.Error(t, err)

	os.Clearenv()
}

func TestSandboxDeleteAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "forbidden"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STOPPED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err := sandbox.doDeleteWithTimeout(ctx, 5*time.Second)
	require.Error(t, err)

	os.Clearenv()
}

func TestSandboxSetLabelsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err := sandbox.doSetLabels(ctx, map[string]string{"env": "dev"})
	require.Error(t, err)

	os.Clearenv()
}

func TestSandboxArchiveAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "archive failed"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	sandbox := NewSandbox(client, nil, "test-id", "test", apiclient.SANDBOXSTATE_STOPPED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err := sandbox.doArchive(ctx)
	require.Error(t, err)

	os.Clearenv()
}
