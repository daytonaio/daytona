// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"context"
	"fmt"
	"net/http"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/google/uuid"
)

// isVolumeId reports whether the argument is a canonical UUID (the length check
// excludes braced/URN/dashless variants).
func isVolumeId(arg string) bool {
	return len(arg) == 36 && uuid.Validate(arg) == nil
}

// resolveVolume resolves a VOLUME_ID_OR_NAME argument to a volume. A UUID-shaped
// argument may be a volume ID or another volume's name, so it is looked up both
// ways; when the lookups find two different volumes the reference is ambiguous
// and resolution fails rather than guessing — the same rule the API applies when
// resolving volume mounts on sandbox create.
func resolveVolume(ctx context.Context, apiClient *apiclient.APIClient, ref string) (*apiclient.VolumeDto, error) {
	var idVolume *apiclient.VolumeDto
	if isVolumeId(ref) {
		volume, res, err := apiClient.VolumesAPI.GetVolume(ctx, ref).Execute()
		if err == nil {
			// Deleted volumes keep their row and stay fetchable by ID; treat them
			// as not found, the same way the by-name lookup does.
			if volume.State != apiclient.VOLUMESTATE_DELETED {
				idVolume = volume
			}
		} else if res == nil || res.StatusCode != http.StatusNotFound {
			return nil, apiclient_cli.HandleErrorResponse(res, err)
		}
	}

	nameVolume, nameRes, nameErr := apiClient.VolumesAPI.GetVolumeByName(ctx, ref).Execute()
	if nameErr != nil && (nameRes == nil || nameRes.StatusCode != http.StatusNotFound) {
		return nil, apiclient_cli.HandleErrorResponse(nameRes, nameErr)
	}

	if idVolume != nil && nameErr == nil && idVolume.Id != nameVolume.Id {
		return nil, fmt.Errorf("volume reference %q matches one volume's ID and another volume's name; rename the volume to remove the ambiguity", ref)
	}
	if idVolume != nil {
		return idVolume, nil
	}
	if nameErr == nil {
		return nameVolume, nil
	}
	return nil, apiclient_cli.HandleErrorResponse(nameRes, nameErr)
}
