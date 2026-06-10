// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common_test

import (
	"errors"
	"testing"

	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	"github.com/spf13/cobra"
)

func TestFormatFlagPreRunEValidation(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "empty value is accepted", value: "", wantErr: false},
		{name: "json is accepted", value: "json", wantErr: false},
		{name: "yaml is accepted", value: "yaml", wantErr: false},
		{name: "xml is rejected", value: "xml", wantErr: true},
		{name: "uppercase JSON is rejected", value: "JSON", wantErr: true},
		{name: "whitespace value is rejected", value: " json", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevFormat := common.FormatFlag
			t.Cleanup(func() {
				common.UnblockStdOut()
				common.FormatFlag = prevFormat
				internal.SuppressVersionMismatchWarning = false
			})

			cmd := &cobra.Command{Use: "test"}
			common.RegisterFormatFlag(cmd)
			if cmd.PreRunE == nil {
				t.Fatal("RegisterFormatFlag did not set PreRunE")
			}

			common.FormatFlag = tt.value
			err := cmd.PreRunE(cmd, nil)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("PreRunE with format %q returned unexpected error: %v", tt.value, err)
				}
				return
			}

			var cliErr *clierr.Error
			if !errors.As(err, &cliErr) {
				t.Fatalf("PreRunE with format %q expected *clierr.Error, got %T: %v", tt.value, err, err)
			}
			if cliErr.Category != clierr.CategoryUsage {
				t.Errorf("PreRunE with format %q category = %q, want %q", tt.value, cliErr.Category, clierr.CategoryUsage)
			}
		})
	}
}

func TestRegisterFormatFlagShorthands(t *testing.T) {
	tests := []struct {
		name          string
		register      func(*cobra.Command)
		wantShorthand string
	}{
		{name: "RegisterFormatFlag keeps -f shorthand", register: common.RegisterFormatFlag, wantShorthand: "f"},
		{name: "RegisterFormatFlagNoShorthand has no shorthand", register: common.RegisterFormatFlagNoShorthand, wantShorthand: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			tt.register(cmd)

			flag := cmd.Flags().Lookup("format")
			if flag == nil {
				t.Fatal("--format flag not registered")
			}
			if flag.Shorthand != tt.wantShorthand {
				t.Errorf("--format shorthand = %q, want %q", flag.Shorthand, tt.wantShorthand)
			}
			if cmd.PreRunE == nil {
				t.Error("PreRunE not set")
			}
		})
	}
}
