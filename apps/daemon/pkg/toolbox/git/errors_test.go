// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"testing"

	"github.com/daytonaio/daemon/pkg/common"
	go_git "github.com/go-git/go-git/v5"
)

func TestClassifyGitError_BranchExistsSentinel(t *testing.T) {
	out := classifyGitError(go_git.ErrBranchExists)
	if _, ok := out.(*common.GitBranchExistsError); !ok {
		t.Fatalf("expected *common.GitBranchExistsError, got %T: %v", out, out)
	}
}

func TestClassifyGitError_BranchNotFound(t *testing.T) {
	err := fmt.Errorf("%w: no-such-branch", go_git.ErrBranchNotFound)
	out := classifyGitError(err)
	if _, ok := out.(*common.GitBranchNotFoundError); !ok {
		t.Fatalf("expected *common.GitBranchNotFoundError, got %T: %v", out, out)
	}
}
