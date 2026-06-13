// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// gitRun runs a git command in dir with hooks/signing disabled, failing the
// test on error.
func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
	base := []string{
		"-C", dir,
		"-c", "core.hooksPath=/dev/null",
		"-c", "commit.gpgSign=false",
		"-c", "tag.gpgSign=false",
	}
	cmd := exec.Command(gitBin, append(base, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
}

// gitCommit creates a commit in dir with the given message.
func gitCommit(t *testing.T, dir, message string) {
	t.Helper()
	gitRun(t, dir, "commit", "-m", message)
}

// gitOutput runs a git command in dir and returns its trimmed stdout.
func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
	cmd := exec.Command(gitBin, append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	require.NoError(t, err, "git %v", args)
	return strings.TrimRight(string(out), "\n")
}

func boolPtr(b bool) *bool { return &b }

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(b)
}

func TestResolveRestoreTargets(t *testing.T) {
	tests := []struct {
		name         string
		staged       *bool
		worktree     *bool
		wantStaged   bool
		wantWorktree bool
	}{
		{"defaults to worktree", nil, nil, false, true},
		{"staged only", boolPtr(true), nil, true, false},
		{"worktree only", nil, boolPtr(true), false, true},
		{"both", boolPtr(true), boolPtr(true), true, true},
		{"explicit false worktree", nil, boolPtr(false), false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			staged, worktree := resolveRestoreTargets(tc.staged, tc.worktree)
			require.Equal(t, tc.wantStaged, staged)
			require.Equal(t, tc.wantWorktree, worktree)
		})
	}
}

func TestSetGetConfig_LocalScope(t *testing.T) {
	dir := t.TempDir()
	svc := Service{WorkDir: dir}

	// Local config lives at <repo>/.git/config (go-git, no git binary needed).
	require.NoError(t, svc.Init(false, "main"))

	// Unset key returns nil.
	value, err := svc.GetConfigValue("user.name", "local")
	require.NoError(t, err)
	require.Nil(t, value)

	require.NoError(t, svc.SetConfigValue("user.name", "Alice", "local"))

	value, err = svc.GetConfigValue("user.name", "local")
	require.NoError(t, err)
	require.NotNil(t, value)
	require.Equal(t, "Alice", *value)

	// ConfigureUser sets both name and email.
	require.NoError(t, svc.ConfigureUser("Bob", "bob@example.com", "local"))

	name, err := svc.GetConfigValue("user.name", "local")
	require.NoError(t, err)
	require.Equal(t, "Bob", *name)

	email, err := svc.GetConfigValue("user.email", "local")
	require.NoError(t, err)
	require.Equal(t, "bob@example.com", *email)
}

func TestInitAndRemotes(t *testing.T) {
	dir := t.TempDir()
	svc := Service{WorkDir: dir}

	require.NoError(t, svc.Init(false, "main"))
	require.DirExists(t, filepath.Join(dir, ".git"))

	require.NoError(t, svc.AddRemote("origin", "https://github.com/user/repo.git", false, false))

	remotes, err := svc.ListRemotes()
	require.NoError(t, err)
	require.Len(t, remotes, 1)
	require.Equal(t, "origin", remotes[0].Name)
	require.Equal(t, "https://github.com/user/repo.git", remotes[0].URL)

	// Adding the same remote without overwrite fails.
	require.Error(t, svc.AddRemote("origin", "https://github.com/user/other.git", false, false))

	// With overwrite it replaces the URL.
	require.NoError(t, svc.AddRemote("origin", "https://github.com/user/other.git", false, true))
	remotes, err = svc.ListRemotes()
	require.NoError(t, err)
	require.Len(t, remotes, 1)
	require.Equal(t, "https://github.com/user/other.git", remotes[0].URL)
}

func TestSetGetConfig_GlobalScope(t *testing.T) {
	// Isolate global config to a temp HOME/XDG so we never touch the real user's
	// ~/.gitconfig.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	svc := Service{}

	value, err := svc.GetConfigValue("user.name", "global")
	require.NoError(t, err)
	require.Nil(t, value)

	require.NoError(t, svc.SetConfigValue("user.name", "Alice", "global"))

	value, err = svc.GetConfigValue("user.name", "global")
	require.NoError(t, err)
	require.NotNil(t, value)
	require.Equal(t, "Alice", *value)
}

func TestAuthenticate(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// Isolate HOME so we write to a temp ~/.gitconfig and ~/.git-credentials.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	svc := Service{}
	require.NoError(t, svc.Authenticate("ci-bot", "ghp_token123", "example.com", "https"))

	// credential.helper=store written to global config (via go-git).
	helper, err := svc.GetConfigValue("credential.helper", "global")
	require.NoError(t, err)
	require.NotNil(t, helper)
	require.Equal(t, "store", *helper)

	// `git credential approve` persisted the credential to the store.
	require.Equal(t,
		"https://ci-bot:ghp_token123@example.com\n",
		readFile(t, filepath.Join(home, ".git-credentials")),
	)
}

func TestReset_KeepModeCLI(t *testing.T) {
	dir := initTestRepoOnBranch(t, "main")
	svc := Service{WorkDir: dir}

	// Add a second commit so there is something to reset away from.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f"), []byte("v2"), 0o644))
	gitRun(t, dir, "add", "f")
	gitCommit(t, dir, "second")

	// keep is the one reset mode go-git lacks → routed to the CLI.
	require.NoError(t, svc.Reset("keep", "HEAD~1", nil))

	content, err := os.ReadFile(filepath.Join(dir, "f"))
	require.NoError(t, err)
	require.Equal(t, "hi", string(content))
}

func TestRestore_WorktreeFromIndex(t *testing.T) {
	dir := initTestRepoOnBranch(t, "main")
	svc := Service{WorkDir: dir}

	target := filepath.Join(dir, "f")

	// Mutate the working tree only; the index still holds the committed "hi".
	require.NoError(t, os.WriteFile(target, []byte("changed"), 0o644))

	require.NoError(t, svc.Restore([]string{"f"}, nil, nil, ""))

	restored, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "hi", string(restored))
}

func TestRestore_PreservesExecutableBit(t *testing.T) {
	dir := initTestRepoOnBranch(t, "main")
	svc := Service{WorkDir: dir}

	script := filepath.Join(dir, "run.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho hi\n"), 0o755))
	gitRun(t, dir, "add", "run.sh")
	gitCommit(t, dir, "add script")

	// Clobber both content and mode.
	require.NoError(t, os.WriteFile(script, []byte("broken\n"), 0o644))

	require.NoError(t, svc.Restore([]string{"run.sh"}, nil, nil, ""))

	require.Equal(t, "#!/bin/sh\necho hi\n", readFile(t, script))
	info, err := os.Stat(script)
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&0o111, "executable bit should be restored")
}

func TestRestore_WorktreeFromSource(t *testing.T) {
	dir := initTestRepoOnBranch(t, "main") // commit 1: f = "hi"
	svc := Service{WorkDir: dir}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "f"), []byte("world"), 0o644))
	gitRun(t, dir, "add", "f")
	gitCommit(t, dir, "second") // commit 2: f = "world"

	// Restore the working tree copy of f from the previous commit.
	require.NoError(t, svc.Restore([]string{"f"}, nil, boolPtr(true), "HEAD~1"))
	require.Equal(t, "hi", readFile(t, filepath.Join(dir, "f")))
}

func TestRestore_StagedFromSource(t *testing.T) {
	dir := initTestRepoOnBranch(t, "main") // commit 1: f = "hi"
	svc := Service{WorkDir: dir}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "f"), []byte("world"), 0o644))
	gitRun(t, dir, "add", "f")
	gitCommit(t, dir, "second") // commit 2: f = "world", HEAD

	// Restore only the index entry for f from the previous commit; the working
	// tree must be left untouched.
	require.NoError(t, svc.Restore([]string{"f"}, boolPtr(true), boolPtr(false), "HEAD~1"))

	// Working tree still holds commit 2's content.
	require.Equal(t, "world", readFile(t, filepath.Join(dir, "f")))

	// The staged (index) blob now matches commit 1's "hi": `git diff --cached`
	// shows f reverting world -> hi.
	gitRun(t, dir, "diff", "--cached", "--name-only") // sanity: f is staged-changed
	staged := gitOutput(t, dir, "show", ":f")         // the staged version of f
	require.Equal(t, "hi", staged)
}

func TestDeleteBranch_Unconditional(t *testing.T) {
	dir := initTestRepoOnBranch(t, "main")
	svc := Service{WorkDir: dir}

	// An unmerged branch is deleted unconditionally.
	gitRun(t, dir, "checkout", "-b", "unmerged")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "g"), []byte("x"), 0o644))
	gitRun(t, dir, "add", "g")
	gitCommit(t, dir, "feature work")
	gitRun(t, dir, "checkout", "main")

	require.NoError(t, svc.DeleteBranch("unmerged"))

	branches, _, err := svc.ListBranches()
	require.NoError(t, err)
	require.NotContains(t, branches, "unmerged")
}

func TestRedactRemoteURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://user:token@github.com/org/repo.git", "https://user@github.com/org/repo.git"},
		{"https://user@github.com/org/repo.git", "https://user@github.com/org/repo.git"},
		{"https://github.com/org/repo.git", "https://github.com/org/repo.git"},
		{"git@github.com:org/repo.git", "git@github.com:org/repo.git"},
		{"", ""},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, redactRemoteURL(tc.in), tc.in)
	}
}

func TestListBranches_CurrentOnFreshRepo(t *testing.T) {
	dir := t.TempDir()
	gitRun(t, dir, "init", "-b", "trunk")
	svc := Service{WorkDir: dir}

	// No commit yet: the branch is unborn, but current must reflect HEAD's target
	// and branches must serialize as [] (non-nil), not null.
	branches, current, err := svc.ListBranches()
	require.NoError(t, err)
	require.NotNil(t, branches)
	require.Empty(t, branches)
	require.Equal(t, "trunk", current)
}
