// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common_test

import (
	"testing"

	"github.com/daytonaio/daytona/cli/cmd/common"
)

func TestShellJoinArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "single argument is returned verbatim",
			args: []string{"echo hi | wc -l"},
			want: "echo hi | wc -l",
		},
		{
			name: "single argument with quotes is returned verbatim",
			args: []string{`echo "hello world" && ls`},
			want: `echo "hello world" && ls`,
		},
		{
			name: "safe arguments stay unquoted",
			args: []string{"ls", "-la", "/tmp/dir", "--format=long", "user@host:path", "a,b+c%d^e"},
			want: "ls -la /tmp/dir --format=long user@host:path a,b+c%d^e",
		},
		{
			name: "argument with spaces is quoted",
			args: []string{"echo", "hello world"},
			want: "echo 'hello world'",
		},
		{
			name: "single quote is escaped",
			args: []string{"echo", "it's"},
			want: `echo 'it'\''s'`,
		},
		{
			name: "double quotes are wrapped in single quotes",
			args: []string{"echo", `say "hi"`},
			want: `echo 'say "hi"'`,
		},
		{
			name: "dollar variable is quoted",
			args: []string{"echo", "$HOME"},
			want: "echo '$HOME'",
		},
		{
			name: "glob is quoted",
			args: []string{"ls", "*.go"},
			want: "ls '*.go'",
		},
		{
			name: "empty argument becomes empty quotes",
			args: []string{"printf", ""},
			want: "printf ''",
		},
		{
			name: "no arguments yields empty string",
			args: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.ShellJoinArgs(tt.args); got != tt.want {
				t.Errorf("ShellJoinArgs(%q) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}
