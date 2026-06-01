// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"os/exec"
	"testing"

	go_git "github.com/go-git/go-git/v5"
)

func initRepoWithCommit(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	for _, args := range [][]string{
		{"init", "-b", "master"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
		{"commit", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func TestCreateBranch_ExistingReturnsSentinel(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
	dir := initRepoWithCommit(t)

	svc := &Service{WorkDir: dir}
	if err := svc.CreateBranch("master"); !errors.Is(err, go_git.ErrBranchExists) {
		t.Fatalf("expected go_git.ErrBranchExists, got %v", err)
	}
}
