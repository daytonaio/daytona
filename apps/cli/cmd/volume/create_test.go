// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import "testing"

func TestCreateCmdHasNoSizeFlag(t *testing.T) {
	if flag := CreateCmd.Flags().Lookup("size"); flag != nil {
		t.Errorf("CreateCmd has a --size flag, want none (the API model has no size field)")
	}
}
