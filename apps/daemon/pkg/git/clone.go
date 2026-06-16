// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/daytonaio/daemon/pkg/gitprovider"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func (s *Service) CloneRepository(repo *gitprovider.GitRepository, auth *http.BasicAuth, insecureSkipTLS bool, depth int) error {
	if isGitCLIModeEnabled() || os.Getenv(experimentalUseGitCloneCLIEnv) == "true" {
		return s.CloneRepositoryCLI(repo, auth, insecureSkipTLS, depth)
	}

	cloneOptions := &git.CloneOptions{
		URL:             repo.Url,
		SingleBranch:    true,
		InsecureSkipTLS: insecureSkipTLS,
		Auth:            auth,
	}

	if depth > 0 {
		cloneOptions.Depth = depth
	}

	if s.LogWriter != nil {
		cloneOptions.Progress = s.LogWriter
	}

	// Azure DevOps requires capabilities multi_ack / multi_ack_detailed,
	// which are not fully implemented and by default are included in
	// transport.UnsupportedCapabilities.
	//
	// This can be removed once go-git implements the git v2 protocol.
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	if repo.Branch != "" {
		cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(repo.Branch)
	}

	_, err := git.PlainClone(s.WorkDir, false, cloneOptions)
	if err != nil {
		return err
	}

	if repo.Target == gitprovider.CloneTargetCommit {
		r, err := git.PlainOpen(s.WorkDir)
		if err != nil {
			return err
		}

		w, err := r.Worktree()
		if err != nil {
			return err
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(repo.Sha),
		})
		if err != nil {
			return err
		}
	}

	return err
}

// CloneRepositoryCLI clones via the `git` CLI. Bounded memory (mmap pack handling).
// Creds flow through GIT_ASKPASS + env — never via URL or argv.
func (s *Service) CloneRepositoryCLI(repo *gitprovider.GitRepository, auth *http.BasicAuth, insecureSkipTLS bool, depth int) error {
	if err := s.runGitCLI(gitCLIOptions{
		op:       "git clone",
		args:     buildCloneArgs(repo, s.WorkDir, insecureSkipTLS, depth),
		auth:     auth,
		tailSize: 64 * 1024,
	}); err != nil {
		return err
	}

	if repo.Target == gitprovider.CloneTargetCommit {
		// Checkout is a purely local op — no network creds needed. Leave auth
		// unset so rogue checkout hooks cannot exfiltrate the token via
		// GIT_USERNAME / GIT_PASSWORD.
		return s.runGitCLI(gitCLIOptions{
			op:       fmt.Sprintf("git checkout %s", repo.Sha),
			args:     buildCheckoutArgs(s.WorkDir, repo.Sha),
			tailSize: 16 * 1024,
		})
	}
	return nil
}

// Credentials must NEVER be embedded in the URL — they flow via GIT_ASKPASS (see buildGitCLIEnv).
// When skipVerify is true, the caller has explicitly opted into insecure TLS via the
// request's insecure_skip_tls flag; we forward that to git via -c http.sslVerify=false.
// When skipVerify is false (default), we do NOT pin sslVerify=true so a sandbox-shell
// user with their own ~/.gitconfig override is still honored on the CLI escape path.
func buildCloneArgs(repo *gitprovider.GitRepository, workDir string, skipVerify bool, depth int) []string {
	cloneURL := repo.Url
	if !strings.Contains(cloneURL, "://") {
		cloneURL = "https://" + cloneURL
	}

	args := []string{
		"-c", "credential.helper=", // prevent any inherited helper from persisting the token
		"-c", "core.hooksPath=/dev/null", // disable post-checkout (and any inherited core.hooksPath) — defense-in-depth so hooks can't read GIT_USERNAME / GIT_PASSWORD
	}
	if skipVerify {
		args = append(args, "-c", "http.sslVerify=false")
	}
	args = append(args,
		"clone",
		"--single-branch",
		"--progress",
	)
	if repo.Branch != "" {
		args = append(args, "--branch", repo.Branch)
	}
	if depth > 0 {
		args = append(args, "--depth", strconv.Itoa(depth))
	}
	args = append(args, "--", cloneURL, workDir)
	return args
}

func buildCheckoutArgs(workDir, sha string) []string {
	// No `--` separator: that would make git treat the SHA as a pathspec.
	return []string{"-C", workDir, "checkout", sha}
}
