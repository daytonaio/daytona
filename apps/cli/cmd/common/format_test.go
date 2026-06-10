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

func TestRegisterFormatFlagChainsPreRunE(t *testing.T) {
	tests := []struct {
		name            string
		value           string
		wantErr         bool
		wantOriginalRun bool
	}{
		{name: "valid format runs original PreRunE after validation", value: "json", wantOriginalRun: true},
		{name: "empty format runs original PreRunE", value: "", wantOriginalRun: true},
		{name: "invalid format short-circuits original PreRunE", value: "xml", wantErr: true, wantOriginalRun: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevFormat := common.FormatFlag
			t.Cleanup(func() {
				common.UnblockStdOut()
				common.FormatFlag = prevFormat
				internal.SuppressVersionMismatchWarning = false
			})

			originalRun := false
			cmd := &cobra.Command{Use: "test"}
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				originalRun = true
				return nil
			}
			common.RegisterFormatFlag(cmd)

			common.FormatFlag = tt.value
			err := cmd.PreRunE(cmd, nil)

			if tt.wantErr && err == nil {
				t.Error("PreRunE expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("PreRunE unexpected error: %v", err)
			}
			if originalRun != tt.wantOriginalRun {
				t.Errorf("original PreRunE run = %v, want %v", originalRun, tt.wantOriginalRun)
			}
		})
	}
}

func TestRegisterFormatFlagChainPropagatesOriginalError(t *testing.T) {
	prevFormat := common.FormatFlag
	t.Cleanup(func() {
		common.UnblockStdOut()
		common.FormatFlag = prevFormat
		internal.SuppressVersionMismatchWarning = false
	})

	wantErr := errors.New("original failed")
	cmd := &cobra.Command{Use: "test"}
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return wantErr }
	common.RegisterFormatFlagNoShorthand(cmd)

	common.FormatFlag = "json"
	if err := cmd.PreRunE(cmd, nil); !errors.Is(err, wantErr) {
		t.Errorf("PreRunE error = %v, want original error %v", err, wantErr)
	}
}

func TestPrintWithoutFormatIsNoOp(t *testing.T) {
	prevFormat := common.FormatFlag
	t.Cleanup(func() { common.FormatFlag = prevFormat })

	// With no --format value the formatter is nil; Print must return
	// without panicking (and without touching stdout blocking state).
	common.FormatFlag = ""
	common.NewFormatter(struct{ Name string }{Name: "x"}).Print()
}
