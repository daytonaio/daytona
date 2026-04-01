// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.CreateSandboxFromSnapshotParams;
import io.daytona.sdk.model.ExecuteResponse;
import io.daytona.sdk.model.SessionExecuteRequest;
import io.daytona.sdk.model.SessionExecuteResponse;

public class ExecCommand {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            System.out.println("Creating sandbox with Python language");
            CreateSandboxFromSnapshotParams params = new CreateSandboxFromSnapshotParams();
            params.setLanguage("python");
            Sandbox sandbox = daytona.create(params);
            System.out.println("Sandbox created: " + sandbox.getId());

            try {
                ExecuteResponse cmd = sandbox.getProcess().executeCommand("echo \"Hello World from CMD!\"");
                System.out.println(cmd.getResult());

                ExecuteResponse code = sandbox.getProcess().codeRun("print(\"Hello World from code!\")");
                System.out.println(code.getResult());

                System.out.println("Creating session");
                sandbox.getProcess().createSession("test-session-1");
                sandbox.getProcess().executeSessionCommand("test-session-1", new SessionExecuteRequest("export FOO=BAR", false));

                SessionExecuteResponse sessionEcho = sandbox.getProcess().executeSessionCommand(
                        "test-session-1",
                        new SessionExecuteRequest("echo $FOO", false)
                );
                System.out.println("FOO=" + sessionEcho.getOutput().trim());

                String logs = sandbox.getProcess().getSessionCommandLogs("test-session-1", sessionEcho.getCmdId());
                System.out.println("Session command logs: " + logs);

                System.out.println("Deleting session");
                sandbox.getProcess().deleteSession("test-session-1");
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
            }
        }
    }
}
