// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.PtyCreateOptions;
import io.daytona.sdk.PtyHandle;
import io.daytona.sdk.PtyResult;
import io.daytona.sdk.Sandbox;

public class Pty {
    public static void main(String[] args) throws Exception {
        try (Daytona daytona = new Daytona()) {
            Sandbox sandbox = daytona.create();
            try {
                System.out.println("=== First PTY Session: Interactive Command with Exit ===");
                PtyHandle handle = sandbox.process.createPty(new PtyCreateOptions(
                        "interactive-pty",
                        120,
                        30,
                        data -> {
                            try {
                                System.out.write(data);
                                System.out.flush();
                            } catch (Exception e) {
                                throw new RuntimeException(e);
                            }
                        }
                ));

                handle.waitForConnection(10);
                System.out.println("\nSending interactive command...");
                handle.sendInput("echo 'Hello from PTY!'\n");
                Thread.sleep(1000);
                handle.sendInput("exit\n");

                PtyResult result = handle.waitForExit(10);
                System.out.println("\nPTY session exited with code: " + result.getExitCode());
                handle.disconnect();

                System.out.println("\n=== Second PTY Session: Kill PTY Session ===");
                PtyHandle handle2 = sandbox.process.createPty(new PtyCreateOptions(
                        "kill-pty",
                        120,
                        30,
                        data -> {}
                ));
                handle2.waitForConnection(10);
                System.out.println("Sending long-running command (infinite loop)...");
                handle2.sendInput("while true; do echo running; sleep 1; done\n");
                Thread.sleep(2000);
                handle2.kill();

                PtyResult result2 = handle2.waitForExit(10);
                System.out.println("\nPTY session terminated. Exit code: " + result2.getExitCode());
                handle2.disconnect();
            } finally {
                System.out.println("\nDeleting sandbox: " + sandbox.getId());
                sandbox.delete();
            }
        }
    }
}
