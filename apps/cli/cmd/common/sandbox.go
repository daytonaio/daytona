// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func RequireStartedState(sandbox *apiclient.Sandbox) error {
	if sandbox.State == nil {
		return fmt.Errorf("sandbox state is unknown")
	}

	state := *sandbox.State
	if state == apiclient.SANDBOXSTATE_STARTED {
		return nil
	}

	sandboxRef := sandbox.Id
	if sandbox.Name != "" {
		sandboxRef = sandbox.Name
	}

	switch state {
	case apiclient.SANDBOXSTATE_STOPPED:
		return fmt.Errorf("sandbox is stopped. Start it with: daytona sandbox start %s", sandboxRef)
	case apiclient.SANDBOXSTATE_ARCHIVED:
		return fmt.Errorf("sandbox is archived. Start it with: daytona sandbox start %s", sandboxRef)
	case apiclient.SANDBOXSTATE_ARCHIVING:
		return fmt.Errorf("sandbox is archiving. Start it with: daytona sandbox start %s", sandboxRef)
	case apiclient.SANDBOXSTATE_STARTING:
		return fmt.Errorf("sandbox is starting. Please wait for it to be ready")
	case apiclient.SANDBOXSTATE_STOPPING:
		return fmt.Errorf("sandbox is stopping. Please wait for it to complete")
	case apiclient.SANDBOXSTATE_CREATING:
		return fmt.Errorf("sandbox is being created. Please wait for it to be ready")
	case apiclient.SANDBOXSTATE_DESTROYING:
		return fmt.Errorf("sandbox is being destroyed")
	case apiclient.SANDBOXSTATE_DESTROYED:
		return fmt.Errorf("sandbox has been destroyed")
	case apiclient.SANDBOXSTATE_ERROR:
		return fmt.Errorf("sandbox is in an error state")
	case apiclient.SANDBOXSTATE_BUILD_FAILED:
		return fmt.Errorf("sandbox build failed")
	default:
		return fmt.Errorf("sandbox is not running (state: %s)", state)
	}
}
