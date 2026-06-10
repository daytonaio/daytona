// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package auth

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
)

func TestResolveAPIKey(t *testing.T) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key")
	if err := os.WriteFile(keyFile, []byte("  file-key\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	emptyFile := filepath.Join(dir, "empty")
	if err := os.WriteFile(emptyFile, []byte(" \n\t"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		flagVal      string
		stdinFlag    bool
		fileFlag     string
		stdin        string
		want         string
		wantErr      bool
		wantUsageErr bool
	}{
		{name: "flag wins over stdin and file", flagVal: "flag-key", stdinFlag: true, fileFlag: keyFile, stdin: "stdin-key", want: "flag-key"},
		{name: "stdin key trimmed", stdinFlag: true, stdin: "  stdin-key \n", want: "stdin-key"},
		{name: "file key trimmed", fileFlag: keyFile, want: "file-key"},
		{name: "empty stdin is usage error", stdinFlag: true, stdin: " \n", wantErr: true, wantUsageErr: true},
		{name: "empty file is usage error", fileFlag: emptyFile, wantErr: true, wantUsageErr: true},
		{name: "missing file errors", fileFlag: filepath.Join(dir, "missing"), wantErr: true},
		{name: "no source resolves to empty key", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveAPIKey(tt.flagVal, tt.stdinFlag, tt.fileFlag, strings.NewReader(tt.stdin))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var cliErr *clierr.Error
				isUsage := errors.As(err, &cliErr) && cliErr.Category == clierr.CategoryUsage
				if isUsage != tt.wantUsageErr {
					t.Errorf("usage-category clierr = %v, want %v (err: %v)", isUsage, tt.wantUsageErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("resolveAPIKey = %q, want %q", got, tt.want)
			}
		})
	}
}
