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

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVolumeServiceCreation(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	vs := NewVolumeService(client)
	require.NotNil(t, vs)

	os.Clearenv()
}

func createTestClientWithServer(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")
	os.Setenv("DAYTONA_API_URL", server.URL)

	client, err := NewClient()
	require.NoError(t, err)
	return client
}

func TestVolumeListError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "internal error"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Volume.List(ctx)
	require.Error(t, err)

	os.Clearenv()
}

func TestVolumeGetError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Volume.Get(ctx, "nonexistent")
	require.Error(t, err)

	os.Clearenv()
}

func TestVolumeDeleteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "forbidden"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	vol := &types.Volume{ID: "vol-1", Name: "my-volume"}
	ctx := context.Background()
	err := client.Volume.Delete(ctx, vol)
	require.Error(t, err)

	os.Clearenv()
}

func TestVolumeErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "volume not found"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Volume.Get(ctx, "nonexistent")
	require.Error(t, err)

	os.Clearenv()
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
