from daytona import CreateSandboxFromImageParams, Daytona, Image, LspCompletionPosition


def section(title: str) -> None:
    print(f"\n=== {title} ===")


def main():
    daytona = Daytona()

    # Custom image with a TypeScript language server (for the LSP showcase) and git.
    sandbox = daytona.create(
        CreateSandboxFromImageParams(
            image=(
                Image.base("ubuntu:25.10").run_commands(
                    "apt-get update && apt-get install -y --no-install-recommends nodejs npm coreutils git",
                    "curl -fsSL https://deb.nodesource.com/setup_20.x | bash -",
                    "apt-get install -y nodejs",
                    "npm install -g ts-node typescript typescript-language-server",
                )
            ),
        ),
        timeout=200,
        on_snapshot_create_logs=print,
    )

    try:
        git = sandbox.git
        proc = sandbox.process
        repo = "demo-repo"

        # ----------------------------- Git operations -----------------------------
        print("git version:", proc.exec("git --version").result.strip())

        section("init")
        git.init(repo, initial_branch="main")
        print("initialized repo at", repo)

        section("configure_user + get_config (local scope)")
        git.configure_user("Ada Lovelace", "ada@example.com", scope="local", path=repo)
        print("user.name  =", git.get_config("user.name", scope="local", path=repo))
        print("user.email =", git.get_config("user.email", scope="local", path=repo))

        section("set_config / get_config (local) + unset key")
        git.set_config("core.editor", "nano", scope="local", path=repo)
        print("core.editor       =", git.get_config("core.editor", scope="local", path=repo))
        print("user.signingkey   =", git.get_config("user.signingkey", scope="local", path=repo), "(unset -> None)")

        section("remote_add / remotes / remote_get")
        git.remote_add(repo, "origin", "https://github.com/panaverse/learn-typescript.git")
        print("remotes        =", [(r.name, r.url) for r in git.remotes(repo).remotes])
        print("remote_get     =", git.remote_get(repo, "origin"))
        print("remote_get(?)  =", git.remote_get(repo, "upstream"), "(missing -> None)")

        section("add / commit")
        sandbox.fs.upload_file(b"line1\n", f"{repo}/a.txt")
        git.add(repo, ["a.txt"])
        commit = git.commit(repo, "first commit", "Ada Lovelace", "ada@example.com")
        print("commit sha =", commit.sha)

        section("branches (current marker)")
        branches = git.branches(repo)
        print("branches =", branches.branches, "| current =", branches.current)

        section("status (detached / upstream / current)")
        status = git.status(repo)
        print(
            f"current_branch={status.current_branch} detached={status.detached} "
            + f"upstream={status.upstream!r} ahead={status.ahead} behind={status.behind} "
            + f"changed_files={len(status.file_status)}"
        )

        section("create_branch + delete_branch")
        git.create_branch(repo, "feature")
        git.checkout_branch(repo, "main")
        git.delete_branch(repo, "feature")
        print("deleted branch 'feature'")

        section("reset (mixed) -> unstage")
        sandbox.fs.upload_file(b"staged\n", f"{repo}/b.txt")
        git.add(repo, ["b.txt"])
        print("staged before reset:", [(f.name, f.staging) for f in git.status(repo).file_status])
        git.reset(repo)
        print("staged after reset :", [(f.name, f.staging) for f in git.status(repo).file_status])

        section("restore (worktree) -> discard local changes")
        sandbox.fs.upload_file(b"corrupted\n", f"{repo}/a.txt")
        print("a.txt before restore:", proc.exec("cat a.txt", cwd=repo).result.strip())
        git.restore(repo, ["a.txt"])
        print("a.txt after restore :", proc.exec("cat a.txt", cwd=repo).result.strip())

        section("reset (keep) -> CLI fallback path")
        sandbox.fs.upload_file(b"v2\n", f"{repo}/a.txt")
        git.add(repo, ["a.txt"])
        _ = git.commit(repo, "second commit", "Ada Lovelace", "ada@example.com")
        git.reset(repo, mode="keep", target="HEAD~1")
        print("a.txt after keep reset to HEAD~1:", proc.exec("cat a.txt", cwd=repo).result.strip())

        section("clone (shallow, depth=1)")
        git.clone("https://github.com/panaverse/learn-typescript", "shallow", branch="master", depth=1)
        count = proc.exec("git rev-list --count HEAD", cwd="shallow").result.strip()
        print("shallow clone commit count (expect 1) =", count)
        print("shallow current branch =", git.status("shallow").current_branch)

        section("pull (remote + branch)")
        git.pull("shallow", remote="origin", branch="master")
        print("pulled origin/master into shallow clone (already up to date)")

        section("dangerously_authenticate")
        git.dangerously_authenticate("ci-bot", "ghp_faketoken", host="example.com")
        print("credential.helper (global) =", git.get_config("credential.helper", scope="global"))
        creds = proc.exec("test -f ~/.git-credentials && echo present || echo missing").result.strip()
        print("~/.git-credentials =", creds)

        print("\nAll git operations exercised successfully ✅")

        # --------------------------------- LSP -----------------------------------
        project_dir = "learn-typescript"

        section("clone project for LSP")
        git.clone("https://github.com/panaverse/learn-typescript", project_dir, "master")
        git.pull(project_dir)

        # Search for the file we want to work on
        matches = sandbox.fs.find_files(project_dir, "var obj1 = new Base();")
        print("Matches:", matches)

        section("LSP: document symbols + completions")
        # Start the language server
        lsp = sandbox.create_lsp_server("typescript", project_dir)
        lsp.start()

        # Notify the language server of the document we want to work on
        lsp.did_open(matches[0].file)

        # Get symbols in the document
        symbols = lsp.document_symbols(matches[0].file)
        print("Symbols:", symbols)

        # Fix the error in the document
        _ = sandbox.fs.replace_in_files([matches[0].file], "var obj1 = new Base();", "var obj1 = new E();")

        # Notify the language server of the document change
        lsp.did_close(matches[0].file)
        lsp.did_open(matches[0].file)

        # Get completions at a specific position
        completions = lsp.completions(matches[0].file, LspCompletionPosition(line=12, character=18))
        print("Completions:", completions)

    except Exception as error:
        print("Error executing example:", error)
        raise
    finally:
        # Cleanup
        daytona.delete(sandbox)


if __name__ == "__main__":
    main()
