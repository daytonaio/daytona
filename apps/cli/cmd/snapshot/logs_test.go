// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"testing"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestIsSnapshotBuildDone(t *testing.T) {
	tests := []struct {
		state      apiclient.SnapshotState
		wantDone   bool
		wantFailed bool
	}{
		{apiclient.SNAPSHOTSTATE_ACTIVE, true, false},
		{apiclient.SNAPSHOTSTATE_INACTIVE, true, false},
		{apiclient.SNAPSHOTSTATE_ERROR, true, true},
		{apiclient.SNAPSHOTSTATE_BUILD_FAILED, true, true},
		{apiclient.SNAPSHOTSTATE_BUILDING, false, false},
		{apiclient.SNAPSHOTSTATE_PENDING, false, false},
		{apiclient.SNAPSHOTSTATE_PULLING, false, false},
		{apiclient.SNAPSHOTSTATE_REMOVING, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			done, failed := isSnapshotBuildDone(tt.state)
			if done != tt.wantDone || failed != tt.wantFailed {
				t.Errorf("isSnapshotBuildDone(%q) = (%v, %v), want (%v, %v)", tt.state, done, failed, tt.wantDone, tt.wantFailed)
			}
		})
	}
}
