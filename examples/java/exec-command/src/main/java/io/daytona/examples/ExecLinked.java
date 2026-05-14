// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.CreateSandboxFromSnapshotParams;
import io.daytona.sdk.model.ExecuteResponse;

public class ExecLinked {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            Sandbox owner = daytona.create();
            System.out.println("Owner sandbox ready: id=" + owner.getId() + " name=" + owner.getName());

            // Linked sandboxes must be ephemeral (autoDeleteInterval=0).
            CreateSandboxFromSnapshotParams followerParams = new CreateSandboxFromSnapshotParams();
            followerParams.setLinkedSandbox(owner.getId());
            followerParams.setAutoDeleteInterval(0);
            Sandbox follower = daytona.create(followerParams);
            System.out.println("Follower sandbox ready: id=" + follower.getId() + " name=" + follower.getName());

            try {
                // Background the http server with nohup, then poll locally until
                // it binds — so the follower's curl below doesn't race startup.
                System.out.println("\nStarting `python3 -m http.server 3000` in owner '" + owner.getName() + "'");
                String startScript = String.join("\n",
                        "set -e",
                        "mkdir -p /tmp/lnk",
                        "echo 'hello from owner' > /tmp/lnk/index.html",
                        "cd /tmp/lnk",
                        "nohup python3 -m http.server 3000 > /tmp/lnk/srv.log 2>&1 &",
                        "for _ in $(seq 1 20); do",
                        "  if curl -sS --max-time 1 http://127.0.0.1:3000/ >/dev/null 2>&1; then",
                        "    echo READY",
                        "    exit 0",
                        "  fi",
                        "  sleep 0.5",
                        "done",
                        "echo 'server failed to start'",
                        "cat /tmp/lnk/srv.log",
                        "exit 1");
                ExecuteResponse startRes = owner.process.executeCommand(startScript, null, null, 30);
                if (startRes.getExitCode() != 0) {
                    throw new RuntimeException("Failed to start server in owner: " + startRes.getResult());
                }
                System.out.println(startRes.getResult().trim());

                // The link network registers the owner under its sandbox name
                // as a DNS alias, so the follower can reach it by name.
                System.out.println("\nReaching '" + owner.getName() + "' from the follower over the link network");
                ExecuteResponse curlRes = follower.process.executeCommand(
                        "curl -sS --max-time 5 http://" + owner.getName() + ":3000/",
                        null, null, 10);
                if (curlRes.getExitCode() != 0) {
                    throw new RuntimeException("Follower could not reach owner: exit=" + curlRes.getExitCode()
                            + " output=" + curlRes.getResult());
                }
                System.out.println("Response from owner: " + curlRes.getResult().trim());
            } finally {
                System.out.println("\nDeleting follower " + follower.getId());
                follower.delete();
                System.out.println("Deleting owner " + owner.getId());
                owner.delete();
            }
        }
    }
}
