import asyncio

from daytona import AsyncDaytona, CreateSandboxFromImageParams, Image, LspCompletionPosition


def section(title: str) -> None:
    print(f"\n=== {title} ===")


async def main():
    async with AsyncDaytona() as daytona:
        # Custom image with a TypeScript language server (for the LSP showcase) and git.
        sandbox = await daytona.create(
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

            # --------------------------- Git operations ---------------------------
            print("git version:", (await proc.exec("git --version")).result.strip())

            section("init")
            await git.init(repo, initial_branch="main")
            print("initialized repo at", repo)

            section("configure_user + get_config (local scope)")
            await git.configure_user("Ada Lovelace", "ada@example.com", scope="local", path=repo)
            print("user.name  =", await git.get_config("user.name", scope="local", path=repo))
            print("user.email =", await git.get_config("user.email", scope="local", path=repo))

            section("set_config / get_config (local) + unset key")
            await git.set_config("core.editor", "nano", scope="local", path=repo)
            print("core.editor       =", await git.get_config("core.editor", scope="local", path=repo))
            print(
                "user.signingkey   =",
                await git.get_config("user.signingkey", scope="local", path=repo),
                "(unset -> None)",
            )

            section("remote_add / remotes / remote_get")
            await git.remote_add(repo, "origin", "https://github.com/panaverse/learn-typescript.git")
            print("remotes        =", [(r.name, r.url) for r in (await git.remotes(repo)).remotes])
            print("remote_get     =", await git.remote_get(repo, "origin"))
            print("remote_get(?)  =", await git.remote_get(repo, "upstream"), "(missing -> None)")

            section("add / commit")
            await sandbox.fs.upload_file(b"line1\n", f"{repo}/a.txt")
            await git.add(repo, ["a.txt"])
            commit = await git.commit(repo, "first commit", "Ada Lovelace", "ada@example.com")
            print("commit sha =", commit.sha)

            section("branches (current marker)")
            branches = await git.branches(repo)
            print("branches =", branches.branches, "| current =", branches.current)

            section("status (detached / upstream / current)")
            status = await git.status(repo)
            print(
                f"current_branch={status.current_branch} detached={status.detached} "
                + f"upstream={status.upstream!r} ahead={status.ahead} behind={status.behind} "
                + f"changed_files={len(status.file_status)}"
            )

            section("create_branch + delete_branch")
            await git.create_branch(repo, "feature")
            await git.checkout_branch(repo, "main")
            await git.delete_branch(repo, "feature")
            print("deleted branch 'feature'")

            section("reset (mixed) -> unstage")
            await sandbox.fs.upload_file(b"staged\n", f"{repo}/b.txt")
            await git.add(repo, ["b.txt"])
            print("staged before reset:", [(f.name, f.staging) for f in (await git.status(repo)).file_status])
            await git.reset(repo)
            print("staged after reset :", [(f.name, f.staging) for f in (await git.status(repo)).file_status])

            section("restore (worktree) -> discard local changes")
            await sandbox.fs.upload_file(b"corrupted\n", f"{repo}/a.txt")
            print("a.txt before restore:", (await proc.exec("cat a.txt", cwd=repo)).result.strip())
            await git.restore(repo, ["a.txt"])
            print("a.txt after restore :", (await proc.exec("cat a.txt", cwd=repo)).result.strip())

            section("reset (keep) -> CLI fallback path")
            await sandbox.fs.upload_file(b"v2\n", f"{repo}/a.txt")
            await git.add(repo, ["a.txt"])
            _ = await git.commit(repo, "second commit", "Ada Lovelace", "ada@example.com")
            await git.reset(repo, mode="keep", target="HEAD~1")
            print("a.txt after keep reset to HEAD~1:", (await proc.exec("cat a.txt", cwd=repo)).result.strip())

            section("clone (shallow, depth=1)")
            await git.clone("https://github.com/panaverse/learn-typescript", "shallow", branch="master", depth=1)
            count = (await proc.exec("git rev-list --count HEAD", cwd="shallow")).result.strip()
            print("shallow clone commit count (expect 1) =", count)
            print("shallow current branch =", (await git.status("shallow")).current_branch)

            section("pull (remote + branch)")
            await git.pull("shallow", remote="origin", branch="master")
            print("pulled origin/master into shallow clone (already up to date)")

            section("dangerously_authenticate")
            await git.dangerously_authenticate("ci-bot", "ghp_faketoken", host="example.com")
            print("credential.helper (global) =", await git.get_config("credential.helper", scope="global"))
            creds = (await proc.exec("test -f ~/.git-credentials && echo present || echo missing")).result.strip()
            print("~/.git-credentials =", creds)

            print("\nAll git operations exercised successfully ✅")

            # ------------------------------- LSP -------------------------------
            project_dir = "learn-typescript"

            section("clone project for LSP")
            await git.clone("https://github.com/panaverse/learn-typescript", project_dir, "master")
            await git.pull(project_dir)

            # Search for the file we want to work on
            matches = await sandbox.fs.find_files(project_dir, "var obj1 = new Base();")
            print("Matches:", matches)

            section("LSP: document symbols + completions")
            # Start the language server
            lsp = sandbox.create_lsp_server("typescript", project_dir)
            await lsp.start()

            # Notify the language server of the document we want to work on
            await lsp.did_open(matches[0].file)

            # Get symbols in the document
            symbols = await lsp.document_symbols(matches[0].file)
            print("Symbols:", symbols)

            # Fix the error in the document
            _ = await sandbox.fs.replace_in_files([matches[0].file], "var obj1 = new Base();", "var obj1 = new E();")

            # Notify the language server of the document change
            await lsp.did_close(matches[0].file)
            await lsp.did_open(matches[0].file)

            # Get completions at a specific position
            completions = await lsp.completions(matches[0].file, LspCompletionPosition(line=12, character=18))
            print("Completions:", completions)

        except Exception as error:
            print("Error executing example:", error)
            raise
        finally:
            # Cleanup
            await daytona.delete(sandbox)


if __name__ == "__main__":
    asyncio.run(main())
