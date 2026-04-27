// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapshotServiceCreation(t *testing.T) {
	t.Setenv("DAYTONA_API_KEY", "test-api-key")
	t.Setenv("DAYTONA_API_URL", "")
	t.Setenv("DAYTONA_JWT_TOKEN", "")
	t.Setenv("DAYTONA_ORGANIZATION_ID", "")

	client, err := NewClient()
	require.NoError(t, err)

	ss := NewSnapshotService(client)
	require.NotNil(t, ss)
}

func TestSnapshotListError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "internal error"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Snapshot.List(ctx, nil, nil)
	require.Error(t, err)
}

func TestSnapshotGetError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Snapshot.Get(ctx, "nonexistent")
	require.Error(t, err)
}

func TestSnapshotDeleteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "forbidden"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	snap := &types.Snapshot{ID: "snap-1", Name: "my-snapshot"}
	ctx := context.Background()
	err := client.Snapshot.Delete(ctx, snap)
	require.Error(t, err)
}

func TestSnapshotErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "snapshot not found"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Snapshot.Get(ctx, "nonexistent")
	require.Error(t, err)
}

func TestMapSnapshotFromAPI(t *testing.T) {
	apiSnapshot := apiclient.NewSnapshotDtoWithDefaults()
	apiSnapshot.SetId("snap-1")
	apiSnapshot.SetName("test-snapshot")
	apiSnapshot.SetState("active")
	apiSnapshot.SetGeneral(false)
	apiSnapshot.SetCpu(4)
	apiSnapshot.SetGpu(0)
	apiSnapshot.SetMem(8)
	apiSnapshot.SetDisk(30)
	apiSnapshot.SetOrganizationId("org-1")
	apiSnapshot.SetImageName("python:3.11")

	snapshot := mapSnapshotFromAPI(apiSnapshot)
	assert.Equal(t, "snap-1", snapshot.ID)
	assert.Equal(t, "test-snapshot", snapshot.Name)
	assert.Equal(t, "active", snapshot.State)
	assert.Equal(t, "org-1", snapshot.OrganizationID)
	assert.Equal(t, "python:3.11", snapshot.ImageName)
	assert.Equal(t, 4, snapshot.CPU)
	assert.Equal(t, 8, snapshot.Memory)
	assert.Equal(t, 30, snapshot.Disk)
}

func TestSnapshotSuccessOperations(t *testing.T) {
	t.Run("list and get map responses", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Query().Get("page") != "" {
				writeJSONResponse(t, w, http.StatusOK, map[string]any{"items": []any{testSnapshotPayload("snap-1", "first", apiclient.SNAPSHOTSTATE_ACTIVE)}, "total": 1, "page": 1, "totalPages": 1})
				return
			}
			writeJSONResponse(t, w, http.StatusOK, testSnapshotPayload("snap-1", "first", apiclient.SNAPSHOTSTATE_ACTIVE))
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		page, limit := 1, 10
		list, err := client.Snapshot.List(context.Background(), &page, &limit)
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		snapshot, err := client.Snapshot.Get(context.Background(), "snap-1")
		require.NoError(t, err)
		assert.Equal(t, "snap-1", snapshot.ID)
	})

	t.Run("create with image streams logs for active snapshot", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost:
				writeJSONResponse(t, w, http.StatusOK, testSnapshotPayload("snap-2", "created", apiclient.SNAPSHOTSTATE_ACTIVE))
			case strings.Contains(r.URL.Path, "build-logs"):
				_, _ = w.Write([]byte("build line 1\nbuild line 2\n"))
			default:
				writeJSONResponse(t, w, http.StatusOK, testSnapshotPayload("snap-2", "created", apiclient.SNAPSHOTSTATE_ACTIVE))
			}
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		snapshot, logChan, err := client.Snapshot.Create(context.Background(), &types.CreateSnapshotParams{Name: "created", Image: "python:3.12"})
		require.NoError(t, err)
		assert.Equal(t, "snap-2", snapshot.ID)
		logs := make([]string, 0, 4)
		for line := range logChan {
			logs = append(logs, line)
		}
		assert.NotEmpty(t, logs)
	})
}

func TestSnapshotLogStreamingHelpers(t *testing.T) {
	t.Run("streamLogsHTTP handles non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		service := client.Snapshot
		err := service.streamLogsHTTP(context.Background(), "snap-1", make(chan string, 1))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code")
	})

	t.Run("processImageContext returns empty without contexts", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		defer server.Close()
		client := createTestClientWithServer(t, server)
		img := Base("python:3.12")
		ctxHashes, err := client.Snapshot.processImageContext(context.Background(), img)
		require.NoError(t, err)
		assert.Empty(t, ctxHashes)
	})

	t.Run("map snapshot preserves optional fields", func(t *testing.T) {
		now := time.Now().UTC()
		sizeVal := float32(42.5)
		size := *apiclient.NewNullableFloat32(&sizeVal)
		errorReason := *apiclient.NewNullableString(nil)
		lastUsedAt := *apiclient.NewNullableTime(nil)
		apiSnapshot := apiclient.NewSnapshotDto("snap-3", false, "mapped", apiclient.SNAPSHOTSTATE_ACTIVE, size, []string{"python"}, 1, 0, 1024, 10, errorReason, now, now, lastUsedAt)
		apiSnapshot.SetOrganizationId("org-9")
		apiSnapshot.SetImageName("python:3.12")
		mapped := mapSnapshotFromAPI(apiSnapshot)
		require.NotNil(t, mapped.Size)
		assert.Equal(t, 42.5, *mapped.Size)
		assert.Equal(t, "org-9", mapped.OrganizationID)
	})
}
