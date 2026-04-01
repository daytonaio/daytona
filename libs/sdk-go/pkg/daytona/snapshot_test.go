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

func TestSnapshotServiceCreation(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	ss := NewSnapshotService(client)
	require.NotNil(t, ss)

	os.Clearenv()
}

func TestSnapshotListError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "internal error"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Snapshot.List(ctx, nil, nil)
	require.Error(t, err)

	os.Clearenv()
}

func TestSnapshotGetError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Snapshot.Get(ctx, "nonexistent")
	require.Error(t, err)

	os.Clearenv()
}

func TestSnapshotDeleteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "forbidden"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	snap := &types.Snapshot{ID: "snap-1", Name: "my-snapshot"}
	ctx := context.Background()
	err := client.Snapshot.Delete(ctx, snap)
	require.Error(t, err)

	os.Clearenv()
}

func TestSnapshotErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "snapshot not found"})
	}))
	defer server.Close()

	client := createTestClientWithServer(t, server)

	ctx := context.Background()
	_, err := client.Snapshot.Get(ctx, "nonexistent")
	require.Error(t, err)

	os.Clearenv()
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
