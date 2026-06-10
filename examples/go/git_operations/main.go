// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func section(t string) { fmt.Printf("\n=== %s ===\n", t) }

func main() {
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	ctx := context.Background()

	// The default snapshot already ships with git, so no custom image is needed.
	sandbox, err := client.Create(ctx, types.SnapshotParams{}, options.WithTimeout(180*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	defer func() {
		if err := sandbox.Delete(ctx); err != nil {
			log.Printf("Failed to delete sandbox: %v", err)
		}
	}()

	git := sandbox.Git
	repo := "demo-repo"

	must := func(err error, what string) {
		if err != nil {
			log.Fatalf("%s: %v", what, err)
		}
	}
	cat := func(file string) string {
		r, err := sandbox.Process.ExecuteCommand(ctx, "cat "+file, options.WithCwd(repo))
		must(err, "cat")
		return strings.TrimSpace(r.Result)
	}
	write := func(path, content string) {
		must(sandbox.FileSystem.UploadFile(ctx, []byte(content), path), "upload "+path)
	}

	v, err := sandbox.Process.ExecuteCommand(ctx, "git --version")
	must(err, "git --version")
	fmt.Println("git version:", strings.TrimSpace(v.Result))

	section("init")
	must(git.Init(ctx, repo, options.WithInitialBranch("main")), "init")
	fmt.Println("initialized repo at", repo)

	section("ConfigureUser + GetConfig (local scope)")
	must(git.ConfigureUser(ctx, "Ada Lovelace", "ada@example.com", options.WithConfigScope("local"), options.WithConfigPath(repo)), "configureUser")
	name, _ := git.GetConfig(ctx, "user.name", options.WithConfigScope("local"), options.WithConfigPath(repo))
	email, _ := git.GetConfig(ctx, "user.email", options.WithConfigScope("local"), options.WithConfigPath(repo))
	fmt.Println("user.name =", name, "| user.email =", email)

	section("SetConfig / GetConfig (local) + unset key")
	must(git.SetConfig(ctx, "core.editor", "nano", options.WithConfigScope("local"), options.WithConfigPath(repo)), "setConfig")
	editor, _ := git.GetConfig(ctx, "core.editor", options.WithConfigScope("local"), options.WithConfigPath(repo))
	missing, _ := git.GetConfig(ctx, "user.signingkey", options.WithConfigScope("local"), options.WithConfigPath(repo))
	fmt.Printf("core.editor = %q | user.signingkey = %q (unset -> empty)\n", editor, missing)

	section("RemoteAdd / Remotes / RemoteGet")
	must(git.RemoteAdd(ctx, repo, "origin", "https://github.com/panaverse/learn-typescript.git"), "remoteAdd")
	remotes, _ := git.Remotes(ctx, repo)
	fmt.Println("remotes =", remotes)
	url, _ := git.RemoteGet(ctx, repo, "origin")
	miss, _ := git.RemoteGet(ctx, repo, "upstream")
	fmt.Printf("remoteGet origin = %q | upstream = %q (missing -> empty)\n", url, miss)

	section("Add / Commit")
	write(repo+"/a.txt", "line1\n")
	must(git.Add(ctx, repo, []string{"a.txt"}), "add")
	commit, err := git.Commit(ctx, repo, "first commit", "Ada Lovelace", "ada@example.com")
	must(err, "commit")
	fmt.Println("commit sha =", commit.SHA)

	section("Branches (current marker)")
	branches, _ := git.Branches(ctx, repo)
	fmt.Println("branches =", branches)

	section("Status (detached / upstream / current)")
	st, err := git.Status(ctx, repo)
	must(err, "status")
	fmt.Printf("current=%s detached=%v upstream=%q ahead=%d behind=%d\n", st.CurrentBranch, st.Detached, st.Upstream, st.Ahead, st.Behind)

	section("CreateBranch + DeleteBranch")
	must(git.CreateBranch(ctx, repo, "feature"), "createBranch")
	must(git.Checkout(ctx, repo, "main"), "checkout main")
	must(git.DeleteBranch(ctx, repo, "feature"), "deleteBranch")
	fmt.Println("deleted branch 'feature'")

	section("Reset (mixed) -> unstage")
	write(repo+"/b.txt", "staged\n")
	must(git.Add(ctx, repo, []string{"b.txt"}), "add b")
	st, _ = git.Status(ctx, repo)
	fmt.Println("staged before reset:", st.FileStatus)
	must(git.Reset(ctx, repo), "reset")
	st, _ = git.Status(ctx, repo)
	fmt.Println("staged after reset :", st.FileStatus)

	section("Restore (worktree) -> discard local changes")
	write(repo+"/a.txt", "corrupted\n")
	fmt.Println("a.txt before restore:", cat("a.txt"))
	must(git.Restore(ctx, repo, []string{"a.txt"}), "restore")
	fmt.Println("a.txt after restore :", cat("a.txt"))

	section("Reset (keep)")
	write(repo+"/a.txt", "v2\n")
	must(git.Add(ctx, repo, []string{"a.txt"}), "add v2")
	_, err = git.Commit(ctx, repo, "second commit", "Ada Lovelace", "ada@example.com")
	must(err, "commit second")
	must(git.Reset(ctx, repo, options.WithResetMode("keep"), options.WithResetTarget("HEAD~1")), "reset keep")
	fmt.Println("a.txt after keep reset to HEAD~1:", cat("a.txt"))

	section("Clone (shallow, depth=1)")
	must(git.Clone(ctx, "https://github.com/panaverse/learn-typescript", "shallow", options.WithBranch("master"), options.WithDepth(1)), "clone")
	count, _ := sandbox.Process.ExecuteCommand(ctx, "git rev-list --count HEAD", options.WithCwd("shallow"))
	fmt.Println("shallow clone commit count (expect 1) =", strings.TrimSpace(count.Result))

	section("Pull (remote + branch)")
	must(git.Pull(ctx, "shallow", options.WithPullRemote("origin"), options.WithPullBranch("master")), "pull")
	fmt.Println("pulled origin/master into shallow clone (already up to date)")

	section("DangerouslyAuthenticate")
	must(git.DangerouslyAuthenticate(ctx, "ci-bot", "ghp_faketoken", options.WithAuthHost("example.com")), "authenticate")
	helper, _ := git.GetConfig(ctx, "credential.helper", options.WithConfigScope("global"))
	fmt.Println("credential.helper (global) =", helper)

	fmt.Println("\nAll new git operations exercised successfully ✅")
}
