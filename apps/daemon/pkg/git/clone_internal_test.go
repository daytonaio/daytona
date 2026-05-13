// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"slices"
	"strings"
	"testing"

	"github.com/daytonaio/daemon/pkg/gitprovider"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/require"
)

// Deliberately distinct from anything that could appear in test URLs, so
// credential-leak regression checks don't false-positive on URL components.
var testCreds = &http.BasicAuth{
	Username: "test-user-xyz",
	Password: "test-token-abc123",
}

func TestBuildCloneArgs(t *testing.T) {
	tests := []struct {
		name     string
		repo     *gitprovider.GitRepository
		workDir  string
		expected []string
	}{
		{
			name: "https URL with branch",
			repo: &gitprovider.GitRepository{
				Url:    "https://github.com/daytonaio/daytona",
				Branch: "main",
			},
			workDir: "/work-dir",
			expected: []string{
				"-c", "credential.helper=",
				"-c", "http.sslVerify=false",
				"clone", "--single-branch", "--progress",
				"--branch", "main",
				"--", "https://github.com/daytonaio/daytona", "/work-dir",
			},
		},
		{
			name: "http URL with branch",
			repo: &gitprovider.GitRepository{
				Url:    "http://localhost:3000/daytonaio/daytona",
				Branch: "main",
			},
			workDir: "/work-dir",
			expected: []string{
				"-c", "credential.helper=",
				"-c", "http.sslVerify=false",
				"clone", "--single-branch", "--progress",
				"--branch", "main",
				"--", "http://localhost:3000/daytonaio/daytona", "/work-dir",
			},
		},
		{
			name: "URL without protocol gets https:// prefix",
			repo: &gitprovider.GitRepository{
				Url:    "github.com/daytonaio/daytona",
				Branch: "main",
			},
			workDir: "/work-dir",
			expected: []string{
				"-c", "credential.helper=",
				"-c", "http.sslVerify=false",
				"clone", "--single-branch", "--progress",
				"--branch", "main",
				"--", "https://github.com/daytonaio/daytona", "/work-dir",
			},
		},
		{
			name: "no branch omits --branch flag",
			repo: &gitprovider.GitRepository{
				Url: "https://github.com/daytonaio/daytona",
			},
			workDir: "/work-dir",
			expected: []string{
				"-c", "credential.helper=",
				"-c", "http.sslVerify=false",
				"clone", "--single-branch", "--progress",
				"--", "https://github.com/daytonaio/daytona", "/work-dir",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildCloneArgs(tc.repo, tc.workDir)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestBuildCloneArgs_NeverEmbedsCredsInURL(t *testing.T) {
	// Regression guard: the old []string-returning implementation embedded
	// user:pass@ in the URL, which persisted into .git/config. The new impl
	// must never do that — creds flow through GIT_ASKPASS env only.
	repo := &gitprovider.GitRepository{
		Url:    "https://github.com/daytonaio/daytona",
		Branch: "main",
	}
	args := buildCloneArgs(repo, "/work-dir")

	for _, arg := range args {
		require.NotContains(t, arg, testCreds.Username, "username leaked into clone args: %q", arg)
		require.NotContains(t, arg, testCreds.Password, "password leaked into clone args: %q", arg)
		require.NotContains(t, arg, "@github.com", "credential-in-URL pattern leaked into clone args: %q", arg)
	}
}

func TestBuildCloneArgs_WithShallowAndAdvancedOptions(t *testing.T) {
	depth := 1
	singleBranch := false
	noTags := true
	sparse := true
	repo := &gitprovider.GitRepository{
		Url:          "https://github.com/daytonaio/daytona",
		Branch:       "main",
		Depth:        &depth,
		SingleBranch: &singleBranch,
		ShallowSince: "2025-01-01",
		NoTags:       &noTags,
		Filter:       "blob:none",
		Sparse:       &sparse,
	}

	got := buildCloneArgs(repo, "/work-dir")

	require.Equal(t, []string{
		"-c", "credential.helper=",
		"-c", "http.sslVerify=false",
		"clone",
		"--no-single-branch",
		"--progress",
		"--depth=1",
		"--shallow-since=2025-01-01",
		"--no-tags",
		"--filter=blob:none",
		"--sparse",
		"--branch", "main",
		"--", "https://github.com/daytonaio/daytona", "/work-dir",
	}, got)
}

func TestBuildCloneArgs_WithReferenceAndSubmoduleOptions(t *testing.T) {
	recurseSubmodules := true
	shallowSubmodules := true
	filterSubmodules := true
	dissociate := true
	repo := &gitprovider.GitRepository{
		Url:               "https://github.com/daytonaio/daytona",
		ReferencePath:     "/cache/daytona.git",
		Dissociate:        &dissociate,
		Filter:            "blob:none",
		RecurseSubmodules: &recurseSubmodules,
		ShallowSubmodules: &shallowSubmodules,
		FilterSubmodules:  &filterSubmodules,
	}

	got := buildCloneArgs(repo, "/work-dir")

	require.Equal(t, []string{
		"-c", "credential.helper=",
		"-c", "http.sslVerify=false",
		"clone",
		"--single-branch",
		"--progress",
		"--filter=blob:none",
		"--reference-if-able=/cache/daytona.git",
		"--dissociate",
		"--recurse-submodules",
		"--shallow-submodules",
		"--also-filter-submodules",
		"--", "https://github.com/daytonaio/daytona", "/work-dir",
	}, got)
}

func TestBuildCloneArgs_WithNoCheckout(t *testing.T) {
	noCheckout := true
	repo := &gitprovider.GitRepository{
		Url:        "https://github.com/daytonaio/daytona",
		NoCheckout: &noCheckout,
	}

	got := buildCloneArgs(repo, "/work-dir")

	require.Equal(t, []string{
		"-c", "credential.helper=",
		"-c", "http.sslVerify=false",
		"clone",
		"--single-branch",
		"--progress",
		"--no-checkout",
		"--", "https://github.com/daytonaio/daytona", "/work-dir",
	}, got)
}

func TestValidateCloneOptions(t *testing.T) {
	depth := 0
	repo := &gitprovider.GitRepository{Depth: &depth}
	require.ErrorContains(t, validateCloneOptions(repo, true), "depth must be greater than or equal to 1")

	depth = 1
	repo = &gitprovider.GitRepository{Depth: &depth}
	require.ErrorContains(t, validateCloneOptions(repo, false), "DAYTONA_EXPERIMENTAL_USE_GIT_CLONE_CLI=true")

	repo = &gitprovider.GitRepository{
		Depth:  &depth,
		Target: gitprovider.CloneTargetCommit,
		Sha:    "abc123",
	}
	require.ErrorContains(t, validateCloneOptions(repo, true), "requires branch to be set")

	repo = &gitprovider.GitRepository{Filter: "blob:none\n--upload-pack=evil"}
	require.ErrorContains(t, validateCloneOptions(repo, true), "filter contains invalid characters")

	repo = &gitprovider.GitRepository{SparsePaths: []string{"../outside"}}
	require.ErrorContains(t, validateCloneOptions(repo, true), "sparse_paths must contain relative paths")

	dissociate := true
	repo = &gitprovider.GitRepository{Dissociate: &dissociate}
	require.ErrorContains(t, validateCloneOptions(repo, true), "dissociate requires reference_path")

	shallowSubmodules := true
	repo = &gitprovider.GitRepository{ShallowSubmodules: &shallowSubmodules}
	require.ErrorContains(t, validateCloneOptions(repo, true), "shallow_submodules requires recurse_submodules")

	filterSubmodules := true
	recurseSubmodules := true
	repo = &gitprovider.GitRepository{FilterSubmodules: &filterSubmodules, RecurseSubmodules: &recurseSubmodules}
	require.ErrorContains(t, validateCloneOptions(repo, true), "filter_submodules requires filter")

	backgroundDeepen := 0
	backgroundExpansion := true
	repo = &gitprovider.GitRepository{BackgroundExpansion: &backgroundExpansion, BackgroundDeepen: &backgroundDeepen}
	require.ErrorContains(t, validateCloneOptions(repo, true), "background_deepen must be greater than or equal to 1")

	repo = &gitprovider.GitRepository{InitialSparsePaths: []string{"src"}}
	require.ErrorContains(t, validateCloneOptions(repo, true), "initial_sparse_paths requires background_expansion")
}

func TestBuildSparseCheckoutArgs(t *testing.T) {
	got := buildSparseCheckoutArgs("/work-dir", []string{"src", "docs/guides"})

	require.Equal(t, []string{
		"-C", "/work-dir",
		"sparse-checkout", "set",
		"--cone",
		"--",
		"src", "docs/guides",
	}, got)
}

func TestBuildBackgroundExpansionArgs(t *testing.T) {
	require.Equal(t, []string{"-C", "/work-dir", "sparse-checkout", "add", "--", "src", "docs"}, buildSparseCheckoutAddArgs("/work-dir", []string{"src", "docs"}))
	require.Equal(t, []string{"-C", "/work-dir", "sparse-checkout", "disable"}, buildSparseCheckoutDisableArgs("/work-dir"))
	require.Equal(t, []string{"-C", "/work-dir", "fetch", "--deepen=50"}, buildFetchDeepenArgs("/work-dir", 50))
	require.Equal(t, []string{"-C", "/work-dir", "fetch", "--unshallow"}, buildFetchUnshallowArgs("/work-dir"))
	require.Equal(t, []string{"-C", "/work-dir", "checkout", "HEAD"}, buildCheckoutHeadArgs("/work-dir"))
	require.Equal(t, []string{"-C", "/work-dir", "checkout", "HEAD", "--", "README.md"}, buildCheckoutPathsArgs("/work-dir", []string{"README.md"}))
	require.Equal(t, []string{"-C", "/work-dir", "maintenance", "run", "--task=prefetch"}, buildMaintenancePrefetchArgs("/work-dir"))
}

func TestBuildCloneEnv_WithCreds(t *testing.T) {
	env := buildCloneEnv(nil, "/tmp/askpass.sh", testCreds)

	require.Contains(t, env, "GIT_ASKPASS=/tmp/askpass.sh")
	require.Contains(t, env, "GIT_TERMINAL_PROMPT=0")
	require.Contains(t, env, "GIT_USERNAME="+testCreds.Username)
	require.Contains(t, env, "GIT_PASSWORD="+testCreds.Password)
}

func TestBuildCloneEnv_WithoutCreds(t *testing.T) {
	env := buildCloneEnv(nil, "/tmp/askpass.sh", nil)

	require.Contains(t, env, "GIT_ASKPASS=/tmp/askpass.sh")
	require.Contains(t, env, "GIT_TERMINAL_PROMPT=0")

	// Without auth, credential env vars must not leak — even as empty values.
	require.False(t, slices.ContainsFunc(env, func(s string) bool {
		return strings.HasPrefix(s, "GIT_USERNAME=")
	}), "GIT_USERNAME must not be set when auth is nil")
	require.False(t, slices.ContainsFunc(env, func(s string) bool {
		return strings.HasPrefix(s, "GIT_PASSWORD=")
	}), "GIT_PASSWORD must not be set when auth is nil")
}

// TestBuildCloneEnv_OverridesBaseEnv covers the glibc getenv-first-match
// quirk: if baseEnv already has any of our managed keys, buildCloneEnv must
// strip them so our values take effect inside the subprocess.
func TestBuildCloneEnv_OverridesBaseEnv(t *testing.T) {
	base := []string{
		"PATH=/usr/bin",
		"GIT_ASKPASS=/wrong/path",
		"GIT_USERNAME=inherited-user",
		"GIT_PASSWORD=inherited-pass",
		"HOME=/home/daytona",
	}
	env := buildCloneEnv(base, "/tmp/askpass.sh", testCreds)

	// Base env unrelated to git is preserved.
	require.Contains(t, env, "PATH=/usr/bin")
	require.Contains(t, env, "HOME=/home/daytona")

	// Managed keys from baseEnv are dropped; our values are the only ones.
	countPrefix := func(prefix string) int {
		n := 0
		for _, kv := range env {
			if strings.HasPrefix(kv, prefix) {
				n++
			}
		}
		return n
	}
	require.Equal(t, 1, countPrefix("GIT_ASKPASS="))
	require.Equal(t, 1, countPrefix("GIT_USERNAME="))
	require.Equal(t, 1, countPrefix("GIT_PASSWORD="))

	require.Contains(t, env, "GIT_ASKPASS=/tmp/askpass.sh")
	require.Contains(t, env, "GIT_USERNAME="+testCreds.Username)
	require.Contains(t, env, "GIT_PASSWORD="+testCreds.Password)
}

func TestBuildCloneEnv_PreservesBaseEnv(t *testing.T) {
	base := []string{"PATH=/usr/bin", "HOME=/home/daytona"}
	env := buildCloneEnv(base, "/tmp/askpass.sh", nil)

	require.Contains(t, env, "PATH=/usr/bin")
	require.Contains(t, env, "HOME=/home/daytona")
}

func TestBuildCheckoutArgs(t *testing.T) {
	got := buildCheckoutArgs("/work-dir", "1234567890")
	// Must NOT include "--" before the SHA — git would then treat it as a
	// pathspec and fail with "did not match any file(s) known to git".
	require.Equal(t, []string{"-C", "/work-dir", "checkout", "1234567890"}, got)
	require.NotContains(t, got, "--", "checkout args must not include -- before the SHA")
}
