// Copyright Daytona Platforms Inc.
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

			sandbox := newSandboxForTest(
				client,
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

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STOPPED, "us-east-1", 60, -1, false, nil)

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

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

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

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STOPPED, "us-east-1", 60, -1, false, nil)

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

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_CREATING, "us-east-1", 60, -1, false, nil)

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

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STOPPING, "us-east-1", 60, -1, false, nil)

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

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

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

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

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

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

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

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

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
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "start failed"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STOPPED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err := sandbox.doStartWithTimeout(ctx, 5*time.Second)
	require.Error(t, err)

	os.Clearenv()
}

func TestSandboxStopAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "stop failed"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err := sandbox.doStopWithTimeout(ctx, 5*time.Second, false)
	require.Error(t, err)

	os.Clearenv()
}

func TestSandboxDeleteAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "forbidden"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STOPPED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err := sandbox.doDeleteWithTimeout(ctx, 5*time.Second)
	require.Error(t, err)

	os.Clearenv()
}

func TestSandboxSetLabelsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STARTED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err := sandbox.doSetLabels(ctx, map[string]string{"env": "dev"})
	require.Error(t, err)

	os.Clearenv()
}

func TestSandboxArchiveAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "archive failed"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	sandbox := newSandboxForTest(client, "test-id", "test", apiclient.SANDBOXSTATE_STOPPED, "us-east-1", 60, -1, false, nil)

	ctx := context.Background()
	err := sandbox.doArchive(ctx)
	require.Error(t, err)

	os.Clearenv()
}

func newSandboxForTest(client *Client, id string, sandboxName string, state apiclient.SandboxState, target string, autoArchiveInterval int, autoDeleteInterval int, networkBlockAll bool, networkAllowList *string) *Sandbox {
	return NewSandbox(client, nil, id, sandboxName, state, target, autoArchiveInterval, autoDeleteInterval, networkBlockAll, networkAllowList, types.CodeLanguagePython)
}

func TestSandboxRefreshDataSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := testSandboxPayload("sb-1", "refreshed", apiclient.SANDBOXSTATE_STARTED)
		payload["target"] = "eu-west-1"
		payload["networkBlockAll"] = true
		payload["networkAllowList"] = "10.0.0.0/8"
		writeJSONResponse(t, w, http.StatusOK, payload)
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)
	sandbox := newSandboxForTest(client, "sb-1", "before", apiclient.SANDBOXSTATE_CREATING, "us-east-1", 1, 2, false, nil)

	require.NoError(t, sandbox.RefreshData(context.Background()))
	assert.Equal(t, "refreshed", sandbox.Name)
	assert.Equal(t, apiclient.SANDBOXSTATE_STARTED, sandbox.State)
	assert.Equal(t, "eu-west-1", sandbox.Target)
	assert.True(t, sandbox.NetworkBlockAll)
	require.NotNil(t, sandbox.NetworkAllowList)
	assert.Equal(t, "10.0.0.0/8", *sandbox.NetworkAllowList)
}

func TestSandboxInfoMethods(t *testing.T) {
	t.Run("successfully gets user home and work dir", func(t *testing.T) {
		var calls int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			switch calls {
			case 1:
				writeJSONResponse(t, w, http.StatusOK, map[string]any{"dir": "/home/daytona"})
			case 2:
				writeJSONResponse(t, w, http.StatusOK, map[string]any{"dir": "/workspace"})
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		sandbox := NewSandbox(nil, createTestToolboxClient(server), "sb", "name", apiclient.SANDBOXSTATE_STARTED, "target", 0, 0, false, nil, types.CodeLanguagePython)
		home, err := sandbox.GetUserHomeDir(context.Background())
		require.NoError(t, err)
		workdir, err := sandbox.GetWorkingDir(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "/home/daytona", home)
		assert.Equal(t, "/workspace", workdir)
	})

	t.Run("converts toolbox errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, http.StatusInternalServerError, map[string]string{"message": "boom"})
		}))
		defer server.Close()

		sandbox := NewSandbox(nil, createTestToolboxClient(server), "sb", "name", apiclient.SANDBOXSTATE_STARTED, "target", 0, 0, false, nil, types.CodeLanguagePython)
		_, err := sandbox.GetUserHomeDir(context.Background())
		require.Error(t, err)
		_, err = sandbox.GetWorkingDir(context.Background())
		require.Error(t, err)
	})
}

func TestSandboxLifecycleSuccessPaths(t *testing.T) {
	t.Run("start stop and delete succeed", func(t *testing.T) {
		var getCount int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				writeJSONResponse(t, w, http.StatusOK, testSandboxPayload("sb", "sandbox", apiclient.SANDBOXSTATE_STARTING))
			case http.MethodGet:
				getCount++
				state := apiclient.SANDBOXSTATE_STARTED
				if getCount > 1 {
					state = apiclient.SANDBOXSTATE_STOPPED
				}
				writeJSONResponse(t, w, http.StatusOK, testSandboxPayload("sb", "sandbox", state))
			case http.MethodDelete:
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		sandbox := newSandboxForTest(client, "sb", "sandbox", apiclient.SANDBOXSTATE_STOPPED, "us", 0, -1, false, nil)
		require.NoError(t, sandbox.doStartWithTimeout(context.Background(), time.Second))
		require.NoError(t, sandbox.doStopWithTimeout(context.Background(), time.Second, true))
		require.NoError(t, sandbox.doDeleteWithTimeout(context.Background(), time.Second))
	})

	t.Run("wait for start returns sandbox error state", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload := testSandboxPayload("sb", "sandbox", apiclient.SANDBOXSTATE_ERROR)
			payload["errorReason"] = "failed"
			writeJSONResponse(t, w, http.StatusOK, payload)
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		sandbox := newSandboxForTest(client, "sb", "sandbox", apiclient.SANDBOXSTATE_STARTING, "us", 0, -1, false, nil)
		err := sandbox.doWaitForStart(context.Background(), 500*time.Millisecond)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Sandbox failed to start")
	})
}

func TestSandboxPreviewAndLabelOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			writeJSONResponse(t, w, http.StatusOK, map[string]any{"labels": map[string]string{"env": "test"}})
		case http.MethodGet:
			writeJSONResponse(t, w, http.StatusOK, testSandboxPayload("sb", "sandbox", apiclient.SANDBOXSTATE_STARTED))
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)
	sandbox := newSandboxForTest(client, "sb", "sandbox", apiclient.SANDBOXSTATE_STARTED, "us", 0, -1, false, nil)
	require.NoError(t, sandbox.SetLabels(context.Background(), map[string]string{"env": "test"}))
	assert.Equal(t, apiclient.SANDBOXSTATE_STARTED, sandbox.State)
}

func TestSandboxExperimentalOperations(t *testing.T) {
	t.Run("fork succeeds and waits for start", func(t *testing.T) {
		var getCount int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				writeJSONResponse(t, w, http.StatusOK, testSandboxPayload("forked", "forked-name", apiclient.SANDBOXSTATE_STARTING))
			case http.MethodGet:
				getCount++
				state := apiclient.SANDBOXSTATE_STARTING
				if getCount > 1 {
					state = apiclient.SANDBOXSTATE_STARTED
				}
				writeJSONResponse(t, w, http.StatusOK, testSandboxPayload("forked", "forked-name", state))
			default:
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		sandbox := newSandboxForTest(client, "sb", "sandbox", apiclient.SANDBOXSTATE_STARTED, "us", 0, -1, false, nil)
		forked, err := sandbox.ExperimentalForkWithTimeout(context.Background(), strPtr("forked-name"), 2*time.Second)
		require.NoError(t, err)
		assert.Equal(t, "forked", forked.ID)
	})

	t.Run("create snapshot waits until snapshotting finishes", func(t *testing.T) {
		var getCount int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				w.WriteHeader(http.StatusOK)
			case http.MethodGet:
				getCount++
				state := apiclient.SANDBOXSTATE_SNAPSHOTTING
				if getCount > 1 {
					state = apiclient.SANDBOXSTATE_STARTED
				}
				writeJSONResponse(t, w, http.StatusOK, testSandboxPayload("sb", "sandbox", state))
			default:
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		sandbox := newSandboxForTest(client, "sb", "sandbox", apiclient.SANDBOXSTATE_STARTED, "us", 0, -1, false, nil)
		require.NoError(t, sandbox.ExperimentalCreateSnapshotWithTimeout(context.Background(), "snap-name", 2*time.Second))
	})
}

func TestSandboxResizeFlow(t *testing.T) {
	var getCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCount++
			state := apiclient.SANDBOXSTATE_RESIZING
			if getCount > 1 {
				state = apiclient.SANDBOXSTATE_STARTED
			}
			writeJSONResponse(t, w, http.StatusOK, testSandboxPayload("sb", "sandbox", state))
		default:
			writeJSONResponse(t, w, http.StatusOK, testSandboxPayload("sb", "sandbox", apiclient.SANDBOXSTATE_RESIZING))
		}
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)
	sandbox := newSandboxForTest(client, "sb", "sandbox", apiclient.SANDBOXSTATE_STARTED, "us", 0, -1, false, nil)
	require.NoError(t, sandbox.ResizeWithTimeout(context.Background(), &types.Resources{CPU: 2, Memory: 2048, Disk: 10}, 2*time.Second))
}
