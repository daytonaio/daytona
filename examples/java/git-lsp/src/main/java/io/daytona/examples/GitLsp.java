// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Image;
import io.daytona.sdk.LspServer;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.CreateSandboxFromImageParams;
import io.daytona.toolbox.client.model.CompletionList;
import io.daytona.toolbox.client.model.LspSymbol;

import java.util.List;
import java.util.Map;

public class GitLsp {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            CreateSandboxFromImageParams params = new CreateSandboxFromImageParams();
            params.setImage(
                    Image.base("ubuntu:25.10").runCommands(
                            "apt-get update && apt-get install -y --no-install-recommends nodejs npm coreutils",
                            "curl -fsSL https://deb.nodesource.com/setup_20.x | bash -",
                            "apt-get install -y nodejs",
                            "npm install -g ts-node typescript typescript-language-server"
                    )
            );
            params.setLanguage("typescript");

            System.out.println("Creating TypeScript sandbox from custom image");
            Sandbox sandbox = daytona.create(params, 300);

            try {
                String projectDir = "learn-typescript";

                System.out.println("Cloning repository");
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
}
