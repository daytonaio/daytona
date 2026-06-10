// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Image;
import io.daytona.sdk.LspServer;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.CreateSandboxFromImageParams;
import io.daytona.sdk.model.GitCommitResponse;
import io.daytona.sdk.model.GitStatus;
import io.daytona.toolbox.client.model.CompletionList;
import io.daytona.toolbox.client.model.LspSymbol;

import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

public class GitLsp {
    static void section(String title) {
        System.out.println("\n=== " + title + " ===");
    }

    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            // Custom image with a TypeScript language server (for the LSP showcase) and git.
            CreateSandboxFromImageParams params = new CreateSandboxFromImageParams();
            params.setImage(
                    Image.base("ubuntu:25.10").runCommands(
                            "apt-get update && apt-get install -y --no-install-recommends nodejs npm coreutils git",
                            "curl -fsSL https://deb.nodesource.com/setup_20.x | bash -",
                            "apt-get install -y nodejs",
                            "npm install -g ts-node typescript typescript-language-server"
                    )
            );
            params.setLanguage("typescript");

            System.out.println("Creating TypeScript sandbox from custom image");
            Sandbox sandbox = daytona.create(params, 300);

            String repo = "demo-repo";
            try {
                var git = sandbox.git;
                var proc = sandbox.process;

                // ----------------------------- Git operations -----------------------------
                System.out.println("git version: " + proc.executeCommand("git --version").getResult().trim());

                section("init");
                git.init(repo, false, "main");
                System.out.println("initialized repo at " + repo);

                section("configureUser + getConfig (local scope)");
                git.configureUser("Ada Lovelace", "ada@example.com", "local", repo);
                System.out.println("user.name  = " + git.getConfig("user.name", "local", repo));
                System.out.println("user.email = " + git.getConfig("user.email", "local", repo));

                section("setConfig / getConfig (local) + unset key");
                git.setConfig("core.editor", "nano", "local", repo);
                System.out.println("core.editor     = " + git.getConfig("core.editor", "local", repo));
                System.out.println("user.signingkey = " + git.getConfig("user.signingkey", "local", repo) + " (unset -> null)");

                section("remoteAdd / remotes / remoteGet");
                git.remoteAdd(repo, "origin", "https://github.com/panaverse/learn-typescript.git");
                System.out.println("remotes       = " + git.remotes(repo).stream()
                        .map(r -> r.getName() + ":" + r.getUrl()).collect(Collectors.toList()));
                System.out.println("remoteGet     = " + git.remoteGet(repo, "origin"));
                System.out.println("remoteGet(?)  = " + git.remoteGet(repo, "upstream") + " (missing -> null)");

                section("add / commit");
                sandbox.fs.uploadFile("line1\n".getBytes(), repo + "/a.txt");
                git.add(repo, List.of("a.txt"));
                GitCommitResponse commit = git.commit(repo, "first commit", "Ada Lovelace", "ada@example.com");
                System.out.println("commit hash = " + commit.getHash());

                section("branches (current marker)");
                System.out.println("branches = " + git.branches(repo).get("branches"));

                section("status (detached / upstream / current)");
                GitStatus s = git.status(repo);
                System.out.printf("current=%s detached=%s upstream=%s ahead=%d behind=%d%n",
                        s.getCurrentBranch(), s.isDetached(), s.getUpstream(), s.getAhead(), s.getBehind());

                section("createBranch + deleteBranch");
                git.createBranch(repo, "feature");
                git.checkoutBranch(repo, "main");
                git.deleteBranch(repo, "feature");
                System.out.println("deleted branch 'feature'");

                section("reset (mixed) -> unstage");
                sandbox.fs.uploadFile("staged\n".getBytes(), repo + "/b.txt");
                git.add(repo, List.of("b.txt"));
                System.out.println("staged before reset: " + fileStatuses(git.status(repo)));
                git.reset(repo);
                System.out.println("staged after reset : " + fileStatuses(git.status(repo)));

                section("restore (worktree) -> discard local changes");
                sandbox.fs.uploadFile("corrupted\n".getBytes(), repo + "/a.txt");
                System.out.println("a.txt before restore: " + proc.executeCommand("cat a.txt", repo, null, null).getResult().trim());
                git.restore(repo, List.of("a.txt"));
                System.out.println("a.txt after restore : " + proc.executeCommand("cat a.txt", repo, null, null).getResult().trim());

                section("reset (keep)");
                sandbox.fs.uploadFile("v2\n".getBytes(), repo + "/a.txt");
                git.add(repo, List.of("a.txt"));
                git.commit(repo, "second commit", "Ada Lovelace", "ada@example.com");
                git.reset(repo, "keep", "HEAD~1", null);
                System.out.println("a.txt after keep reset to HEAD~1: " + proc.executeCommand("cat a.txt", repo, null, null).getResult().trim());

                section("clone (shallow, depth=1)");
                git.clone("https://github.com/panaverse/learn-typescript", "shallow", "master", null, null, null, null, 1);
                System.out.println("shallow clone commit count (expect 1) = "
                        + proc.executeCommand("git rev-list --count HEAD", "shallow", null, null).getResult().trim());

                section("pull (remote + branch)");
                git.pull("shallow", null, null, "master", "origin");
                System.out.println("pulled origin/master into shallow clone (already up to date)");

                section("dangerouslyAuthenticate");
                git.dangerouslyAuthenticate("ci-bot", "ghp_faketoken", "example.com", null);
                System.out.println("credential.helper (global) = " + git.getConfig("credential.helper", "global", null));

                System.out.println("\nAll new git operations exercised successfully ✅");

                // --------------------------------- LSP -----------------------------------
                String projectDir = "learn-typescript";

                section("clone project for LSP");
                sandbox.git.clone("https://github.com/panaverse/learn-typescript", projectDir, "master", null, null, null);

                Map<String, Object> files = sandbox.fs.searchFiles(projectDir, "*.ts");
                System.out.println("TypeScript files: " + files.get("files"));

                List<Map<String, Object>> matches = sandbox.fs.findFiles(projectDir, "var obj1 = new Base();");
                System.out.println("Matches: " + matches);
                if (matches.isEmpty() || matches.get(0).get("file") == null) {
                    System.out.println("No matching file found");
                    return;
                }

                String filePath = String.valueOf(matches.get(0).get("file"));

                section("LSP: document symbols + completions");
                LspServer lsp = sandbox.createLspServer("typescript", projectDir);
                lsp.start("typescript", projectDir);

                lsp.didOpen("typescript", projectDir, filePath);

                List<LspSymbol> symbols = lsp.documentSymbols("typescript", projectDir, filePath);
                System.out.println("Symbols: " + symbols);

                sandbox.fs.replaceInFiles(List.of(filePath), "var obj1 = new Base();", "var obj1 = new E();");

                lsp.didClose("typescript", projectDir, filePath);
                lsp.didOpen("typescript", projectDir, filePath);

                CompletionList completions = lsp.completions("typescript", projectDir, filePath, 12, 18);
                System.out.println("Completions: " + completions);

                lsp.didClose("typescript", projectDir, filePath);
                lsp.stop("typescript", projectDir);
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
            }
        }
    }

    static List<String> fileStatuses(GitStatus s) {
        return s.getFileStatus().stream()
                .map(f -> f.getPath() + ":" + f.getStatus())
                .collect(Collectors.toList());
    }
}
