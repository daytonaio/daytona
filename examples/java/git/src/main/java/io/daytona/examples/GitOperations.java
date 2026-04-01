// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.GitStatus;

import java.util.List;
import java.util.Map;

public class GitOperations {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            System.out.println("Creating sandbox");
            Sandbox sandbox = daytona.create();
            System.out.println("Sandbox created: " + sandbox.getId());

            try {
                String repoPath = "learn-typescript";
                System.out.println("Cloning repository to " + repoPath);
                sandbox.getGit().clone("https://github.com/panaverse/learn-typescript", repoPath);

                Map<String, Object> branches = sandbox.getGit().branches(repoPath);
                Object branchList = branches.get("branches");
                System.out.println("Branches: " + branchList);

                GitStatus status = sandbox.getGit().status(repoPath);
                System.out.println("Current branch: " + status.getCurrentBranch());
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
            }
        }
    }
}
