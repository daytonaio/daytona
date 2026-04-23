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

func TestVolumeServiceCreation(t *testing.T) {
	t.Setenv("DAYTONA_API_KEY", "test-api-key")
	t.Setenv("DAYTONA_API_URL", "")
	t.Setenv("DAYTONA_JWT_TOKEN", "")
	t.Setenv("DAYTONA_ORGANIZATION_ID", "")

	client, err := NewClient()
	require.NoError(t, err)

	vs := NewVolumeService(client)
	require.NotNil(t, vs)
}

func createTestClientWithServer(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	t.Setenv("DAYTONA_API_KEY", "test-api-key")
	t.Setenv("DAYTONA_API_URL", server.URL)
	t.Setenv("DAYTONA_JWT_TOKEN", "")
	t.Setenv("DAYTONA_ORGANIZATION_ID", "")

	client, err := NewClient()
	require.NoError(t, err)
	return client
}

func TestVolumeListError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "internal error"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Volume.List(ctx)
	require.Error(t, err)
}

func TestVolumeGetError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Volume.Get(ctx, "nonexistent")
	require.Error(t, err)
}

func TestVolumeDeleteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "forbidden"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	vol := &types.Volume{ID: "vol-1", Name: "my-volume"}
	ctx := context.Background()
	err := client.Volume.Delete(ctx, vol)
	require.Error(t, err)
}

func TestVolumeErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "volume not found"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Volume.Get(ctx, "nonexistent")
	require.Error(t, err)
}

func TestVolumeDtoToVolume(t *testing.T) {
	dto := apiclient.NewVolumeDtoWithDefaults()
	dto.SetId("vol-1")
	dto.SetName("my-volume")
	dto.SetOrganizationId("org-1")
	dto.SetState(apiclient.VOLUMESTATE_READY)
	dto.SetCreatedAt("2025-01-01T00:00:00Z")
	dto.SetUpdatedAt("2025-01-02T00:00:00Z")

	volume := volumeDtoToVolume(dto)
	assert.Equal(t, "vol-1", volume.ID)
	assert.Equal(t, "my-volume", volume.Name)
	assert.Equal(t, "org-1", volume.OrganizationID)
	assert.Equal(t, "ready", volume.State)
}

func TestVolumeSuccessOperations(t *testing.T) {
	t.Run("list get create and delete succeed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				if strings.Contains(r.URL.Path, "/volumes/") {
					writeJSONResponse(t, w, http.StatusOK, testVolumePayload("vol-1", "my-volume", apiclient.VOLUMESTATE_READY))
					return
				}
				writeJSONResponse(t, w, http.StatusOK, []any{testVolumePayload("vol-1", "my-volume", apiclient.VOLUMESTATE_READY)})
			case http.MethodPost:
				writeJSONResponse(t, w, http.StatusOK, testVolumePayload("vol-1", "my-volume", apiclient.VOLUMESTATE_PENDING_CREATE))
			case http.MethodDelete:
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		volumes, err := client.Volume.List(context.Background())
		require.NoError(t, err)
		assert.Len(t, volumes, 1)
		volume, err := client.Volume.Get(context.Background(), "my-volume")
		require.NoError(t, err)
		assert.Equal(t, "my-volume", volume.Name)
		created, err := client.Volume.Create(context.Background(), "my-volume")
		require.NoError(t, err)
		assert.Equal(t, "pending_create", created.State)
		require.NoError(t, client.Volume.Delete(context.Background(), &types.Volume{ID: "vol-1", Name: "my-volume"}))
	})
}

func TestVolumeWaitForReadyBehaviors(t *testing.T) {
	t.Run("returns ready volume", func(t *testing.T) {
		var calls int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			state := apiclient.VOLUMESTATE_PENDING_CREATE
			if calls > 1 {
				state = apiclient.VOLUMESTATE_READY
			}
			writeJSONResponse(t, w, http.StatusOK, testVolumePayload("vol-1", "my-volume", state))
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		volume, err := client.Volume.WaitForReady(context.Background(), &types.Volume{Name: "my-volume"}, 1500*time.Millisecond)
		require.NoError(t, err)
		assert.Equal(t, "ready", volume.State)
	})

	t.Run("returns error state reason", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload := testVolumePayload("vol-1", "my-volume", apiclient.VOLUMESTATE_ERROR)
			payload["errorReason"] = "quota exceeded"
			writeJSONResponse(t, w, http.StatusOK, payload)
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		_, err := client.Volume.WaitForReady(context.Background(), &types.Volume{Name: "my-volume"}, time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "quota exceeded")
	})

	t.Run("times out", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, testVolumePayload("vol-1", "my-volume", apiclient.VOLUMESTATE_PENDING_CREATE))
		}))
		defer server.Close()

		client := createTestClientWithServer(t, server)
		_, err := client.Volume.WaitForReady(context.Background(), &types.Volume{Name: "my-volume"}, 10*time.Millisecond)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "volume did not become ready")
	})
}
