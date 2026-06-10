// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"testing"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestIsSandboxBuildDone(t *testing.T) {
	tests := []struct {
		state      apiclient.SandboxState
		wantDone   bool
		wantFailed bool
	}{
		{apiclient.SANDBOXSTATE_STARTED, true, false},
		{apiclient.SANDBOXSTATE_STOPPED, true, false},
		{apiclient.SANDBOXSTATE_STOPPING, true, false},
		{apiclient.SANDBOXSTATE_ARCHIVED, true, false},
		{apiclient.SANDBOXSTATE_ARCHIVING, true, false},
		{apiclient.SANDBOXSTATE_DESTROYED, true, false},
		{apiclient.SANDBOXSTATE_DESTROYING, true, false},
		{apiclient.SANDBOXSTATE_ERROR, true, true},
		{apiclient.SANDBOXSTATE_BUILD_FAILED, true, true},
		{apiclient.SANDBOXSTATE_CREATING, false, false},
		{apiclient.SANDBOXSTATE_PENDING_BUILD, false, false},
		{apiclient.SANDBOXSTATE_BUILDING_SNAPSHOT, false, false},
		{apiclient.SANDBOXSTATE_PULLING_SNAPSHOT, false, false},
		{apiclient.SANDBOXSTATE_STARTING, false, false},
		{apiclient.SANDBOXSTATE_RESTORING, false, false},
		{apiclient.SANDBOXSTATE_UNKNOWN, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			done, failed := isSandboxBuildDone(tt.state)
			if done != tt.wantDone || failed != tt.wantFailed {
				t.Errorf("isSandboxBuildDone(%q) = (%v, %v), want (%v, %v)", tt.state, done, failed, tt.wantDone, tt.wantFailed)
			}
		})
	}
}
