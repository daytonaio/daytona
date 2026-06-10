// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
)

func TestExecFailure(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantCategory clierr.Category
		wantMessage  string
		wantHint     string
	}{
		{
			name:         "clierr passthrough keeps category and hint",
			err:          clierr.New(clierr.CategoryNotFound, "sandbox not found").WithHint("check the sandbox ID"),
			wantCategory: clierr.CategoryNotFound,
			wantMessage:  "sandbox not found",
			wantHint:     "check the sandbox ID",
		},
		{
			name:         "wrapped clierr is unwrapped and keeps category",
			err:          fmt.Errorf("toolbox: %w", clierr.New(clierr.CategoryAuth, "unauthorized")),
			wantCategory: clierr.CategoryAuth,
			wantMessage:  "unauthorized",
		},
		{
			name:         "plain error becomes server-category clierr",
			err:          errors.New("connection reset"),
			wantCategory: clierr.CategoryServer,
			wantMessage:  "connection reset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := execFailure(tt.err)

			var cliErr *clierr.Error
			if !errors.As(got, &cliErr) {
				t.Fatalf("execFailure(%v) = %T, want *clierr.Error", tt.err, got)
			}
			if cliErr.Code != 255 {
				t.Errorf("Code = %d, want 255", cliErr.Code)
			}
			if cliErr.Category != tt.wantCategory {
				t.Errorf("Category = %q, want %q", cliErr.Category, tt.wantCategory)
			}
			if cliErr.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", cliErr.Message, tt.wantMessage)
			}
			if cliErr.Hint != tt.wantHint {
				t.Errorf("Hint = %q, want %q", cliErr.Hint, tt.wantHint)
			}
			if exitCode := clierr.ExitCode(got); exitCode != 255 {
				t.Errorf("ExitCode = %d, want 255", exitCode)
			}
		})
	}
}

func TestExecResultTags(t *testing.T) {
	typ := reflect.TypeOf(execResult{})

	tests := []struct {
		field string
		tag   string
	}{
		{field: "Result", tag: "result"},
		{field: "ExitCode", tag: "exitCode"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			f, ok := typ.FieldByName(tt.field)
			if !ok {
				t.Fatalf("execResult has no field %s", tt.field)
			}
			if got := f.Tag.Get("json"); got != tt.tag {
				t.Errorf("json tag = %q, want %q", got, tt.tag)
			}
			if got := f.Tag.Get("yaml"); got != tt.tag {
				t.Errorf("yaml tag = %q, want %q", got, tt.tag)
			}
		})
	}
}
