// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import "testing"

func TestListCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args accepted", args: nil},
		{name: "one arg rejected", args: []string{"extra"}, wantErr: true},
		{name: "two args rejected", args: []string{"extra", "more"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ListCmd.Args(ListCmd, tt.args)
			if tt.wantErr && err == nil {
				t.Fatalf("ListCmd.Args(%v) expected error, got nil", tt.args)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ListCmd.Args(%v) unexpected error: %v", tt.args, err)
			}
		})
	}
}
