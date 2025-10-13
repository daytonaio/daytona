// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

// VolumeService provides persistent storage volume management operations.
//
// VolumeService enables creating, managing, and deleting persistent storage volumes
// that can be attached to sandboxes. Volumes persist data independently of sandbox
// lifecycle and can be shared between sandboxes. Access through [Client.Volumes].
//
// Example:
//
//	// Create a new volume
//	volume, err := client.Volumes.Create(ctx, "my-data-volume")
//	if err != nil {
//	    return err
//	}
//
//	// Wait for volume to be ready
//	volume, err = client.Volumes.WaitForReady(ctx, volume, 60*time.Second)
//	if err != nil {
//	    return err
//	}
//
//	// List all volumes
//	volumes, err := client.Volumes.List(ctx)
type VolumeService struct {
	client *Client
	otel   *otelState
}

// NewVolumeService creates a new VolumeService.
//
// This is typically called internally by the SDK when creating a [Client].
// Users should access VolumeService through [Client.Volumes] rather than
// creating it directly.
func NewVolumeService(client *Client) *VolumeService {
	return &VolumeService{
		client: client,
		otel:   client.Otel,
	}
}

// List returns all volumes in the organization.
//
// Example:
//
//	volumes, err := client.Volumes.List(ctx)
//	if err != nil {
//	    return err
//	}
//	for _, vol := range volumes {
//	    fmt.Printf("Volume %s: %s\n", vol.Name, vol.State)
//	}
//
// Returns a slice of [types.Volume] or an error if the request fails.
func (v *VolumeService) List(ctx context.Context) ([]*types.Volume, error) {
	return withInstrumentation(ctx, v.otel, "Volume", "List", func(ctx context.Context) ([]*types.Volume, error) {
		authCtx := v.client.getAuthContext(ctx)
		volumeDtos, httpResp, err := v.client.apiClient.VolumesAPI.ListVolumes(authCtx).Execute()
		if err != nil {
			return nil, errors.ConvertAPIError(err, httpResp)
		}

		// Convert VolumeDto to types.Volume
		volumes := make([]*types.Volume, len(volumeDtos))
		for i, dto := range volumeDtos {
			volumes[i] = volumeDtoToVolume(&dto)
		}

		return volumes, nil
	})
}

// Get retrieves a volume by its name.
//
// Parameters:
//   - name: The volume name
//
// Example:
//
//	volume, err := client.Volumes.Get(ctx, "my-data-volume")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Volume state: %s\n", volume.State)
//
// Returns the [types.Volume] or an error if not found.
func (v *VolumeService) Get(ctx context.Context, name string) (*types.Volume, error) {
	return withInstrumentation(ctx, v.otel, "Volume", "Get", func(ctx context.Context) (*types.Volume, error) {
		authCtx := v.client.getAuthContext(ctx)
		volumeDto, httpResp, err := v.client.apiClient.VolumesAPI.GetVolumeByName(authCtx, name).Execute()
		if err != nil {
			return nil, errors.ConvertAPIError(err, httpResp)
		}

		return volumeDtoToVolume(volumeDto), nil
	})
}

// Create creates a new persistent storage volume.
//
// The volume starts in "pending" state and transitions to "ready" when available.
// Use [VolumeService.WaitForReady] to wait for the volume to become ready.
//
// Parameters:
//   - name: Unique name for the volume
//
// Example:
//
//	volume, err := client.Volumes.Create(ctx, "my-data-volume")
//	if err != nil {
//	    return err
//	}
//
//	// Wait for volume to be ready
//	volume, err = client.Volumes.WaitForReady(ctx, volume, 60*time.Second)
//
// Returns the created [types.Volume] or an error.
func (v *VolumeService) Create(ctx context.Context, name string) (*types.Volume, error) {
	return withInstrumentation(ctx, v.otel, "Volume", "Create", func(ctx context.Context) (*types.Volume, error) {
		authCtx := v.client.getAuthContext(ctx)

		req := apiclient.NewCreateVolume(name)
		volumeDto, httpResp, err := v.client.apiClient.VolumesAPI.CreateVolume(authCtx).CreateVolume(*req).Execute()
		if err != nil {
			return nil, errors.ConvertAPIError(err, httpResp)
		}

		return volumeDtoToVolume(volumeDto), nil
	})
}

// Delete permanently removes a volume and all its data.
//
// This operation is irreversible. Ensure no sandboxes are using the volume
// before deletion.
//
// Parameters:
//   - volume: The volume to delete
//
// Example:
//
//	err := client.Volumes.Delete(ctx, volume)
//	if err != nil {
//	    return err
//	}
//
// Returns an error if deletion fails.
func (v *VolumeService) Delete(ctx context.Context, volume *types.Volume) error {
	return withInstrumentationVoid(ctx, v.otel, "Volume", "Delete", func(ctx context.Context) error {
		authCtx := v.client.getAuthContext(ctx)
		httpResp, err := v.client.apiClient.VolumesAPI.DeleteVolume(authCtx, volume.ID).Execute()
		if err != nil {
			return errors.ConvertAPIError(err, httpResp)
		}

		return nil
	})
}

// WaitForReady waits for a volume to reach the "ready" state.
//
// This method polls the volume status until it becomes ready, reaches an error state,
// or the timeout expires. The polling interval is 1 second.
//
// Parameters:
//   - volume: The volume to wait for
//   - timeout: Maximum time to wait for the volume to become ready
//
// Example:
//
//	volume, err := client.Volumes.Create(ctx, "my-volume")
//	if err != nil {
//	    return err
//	}
//
//	// Wait up to 2 minutes for the volume to be ready
//	volume, err = client.Volumes.WaitForReady(ctx, volume, 2*time.Minute)
//	if err != nil {
//	    return fmt.Errorf("volume failed to become ready: %w", err)
//	}
//
// Returns the updated [types.Volume] when ready, or an error if the timeout
// expires or the volume enters an error state.
func (v *VolumeService) WaitForReady(ctx context.Context, volume *types.Volume, timeout time.Duration) (*types.Volume, error) {
	return withInstrumentation(ctx, v.otel, "Volume", "WaitForReady", func(ctx context.Context) (*types.Volume, error) {
		deadline := time.Now().Add(timeout)

		for {
			// Check if context is done
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			// Check timeout
			if time.Now().After(deadline) {
				return nil, errors.NewDaytonaTimeoutError("volume did not become ready within timeout")
			}

			// Get current volume state
			currentVolume, err := v.Get(ctx, volume.Name)
			if err != nil {
				return nil, err
			}

			// Check if volume is ready
			if currentVolume.State == "ready" {
				return currentVolume, nil
			}

			// Check if volume is in error state
			if currentVolume.State == "error" {
				errMsg := "volume creation failed"
				if currentVolume.ErrorReason != nil {
					errMsg = *currentVolume.ErrorReason
				}
				return nil, errors.NewDaytonaError(errMsg, 0, nil)
			}

			// Wait before polling again
			time.Sleep(1 * time.Second)
		}
	})
}

// volumeDtoToVolume converts api-client VolumeDto to SDK types.Volume
func volumeDtoToVolume(dto *apiclient.VolumeDto) *types.Volume {
	// Parse timestamps
	createdAt, _ := time.Parse(time.RFC3339, dto.GetCreatedAt())
	updatedAt, _ := time.Parse(time.RFC3339, dto.GetUpdatedAt())

	volume := &types.Volume{
		ID:             dto.GetId(),
		Name:           dto.GetName(),
		OrganizationID: dto.GetOrganizationId(),
		State:          string(dto.GetState()), // Convert VolumeState enum to string
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}

	// Handle nullable LastUsedAt
	if dto.HasLastUsedAt() {
		lastUsedAt, _ := time.Parse(time.RFC3339, dto.GetLastUsedAt())
		volume.LastUsedAt = lastUsedAt
	}

	// Handle nullable ErrorReason
	if dto.ErrorReason.IsSet() {
		volume.ErrorReason = dto.ErrorReason.Get()
	}

	return volume
}
