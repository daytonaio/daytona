// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.CodeInterpreter;
import io.daytona.sdk.Daytona;
import io.daytona.sdk.RunCodeOptions;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.exception.DaytonaTimeoutException;
import io.daytona.sdk.model.CreateSandboxFromSnapshotParams;
import io.daytona.sdk.model.ExecuteResponse;
import io.daytona.sdk.model.SessionExecuteRequest;
import io.daytona.sdk.model.SessionExecuteResponse;
import io.daytona.sdk.model.Session;
import io.daytona.sdk.model.Command;

public class ExecCommand {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            System.out.println("Creating sandbox with Python language");
            CreateSandboxFromSnapshotParams params = new CreateSandboxFromSnapshotParams();
            params.setLanguage("python");
            Sandbox sandbox = daytona.create(params);
            System.out.println("Sandbox created: " + sandbox.getId());

            try {
                basicExec(sandbox);
                sessionExec(sandbox);
                sessionExecLogsAsync(sandbox);
                statefulCodeInterpreter(sandbox);
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
            }
        }
    }

    static void basicExec(Sandbox sandbox) {
        System.out.println("\n" + "=".repeat(60));
        System.out.println("Basic Execution");
        System.out.println("=".repeat(60));

        ExecuteResponse cmd = sandbox.process.executeCommand("echo 'Hello World from CMD!'");
        System.out.println(cmd.getResult());

        ExecuteResponse code = sandbox.process.codeRun("print('Hello World from code!')");
        System.out.println(code.getResult());
    }

    static void sessionExec(Sandbox sandbox) {
        System.out.println("\n" + "=".repeat(60));
        System.out.println("Session Execution");
        System.out.println("=".repeat(60));

        sandbox.process.createSession("exec-session-1");

        Session session = sandbox.process.getSession("exec-session-1");
        System.out.println("Session: " + session.getSessionId());

        SessionExecuteResponse exportCmd = sandbox.process.executeSessionCommand(
                "exec-session-1", new SessionExecuteRequest("export FOO=BAR", false));

        Session sessionUpdated = sandbox.process.getSession("exec-session-1");
        System.out.println("Session commands count: " + sessionUpdated.getCommands().size());

        Command sessionCommand = sandbox.process.getSessionCommand("exec-session-1", exportCmd.getCmdId());
        System.out.println("Command: " + sessionCommand.getCommand());

        SessionExecuteResponse echoCmd = sandbox.process.executeSessionCommand(
                "exec-session-1", new SessionExecuteRequest("echo $FOO", false));
        System.out.println("FOO=" + echoCmd.getOutput().trim());

        var logs = sandbox.process.getSessionCommandLogs("exec-session-1", echoCmd.getCmdId());
        System.out.println("Session command logs: " + logs.getStdout());

        sandbox.process.deleteSession("exec-session-1");
    }

    static void sessionExecLogsAsync(Sandbox sandbox) {
        System.out.println("\n" + "=".repeat(60));
        System.out.println("Async Session with Streaming Logs + Input");
        System.out.println("=".repeat(60));

        String sessionId = "exec-session-async-logs";
        sandbox.process.createSession(sessionId);

        SessionExecuteResponse command = sandbox.process.executeSessionCommand(sessionId,
                new SessionExecuteRequest(
                        "printf 'Enter your name: \\n' && read name && printf 'Hello, %s\\n' \"$name\"; " +
                                "counter=1; while (( counter <= 3 )); do echo \"Count: $counter\"; " +
                                "((counter++)); sleep 2; done",
                        true));

        try { Thread.sleep(1000); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }

        System.out.println("Sending input to the command...");
        sandbox.process.sendSessionCommandInput(sessionId, command.getCmdId(), "Alice\n");
        System.out.println("Input sent");

        sandbox.process.getSessionCommandLogs(sessionId, command.getCmdId(),
                stdout -> System.out.print("[STDOUT]: " + stdout),
                stderr -> System.out.print("[STDERR]: " + stderr));

        sandbox.process.deleteSession(sessionId);
    }

    static void statefulCodeInterpreter(Sandbox sandbox) {
        System.out.println("\n" + "=".repeat(60));
        System.out.println("Stateful Code Interpreter");
        System.out.println("=".repeat(60));

        CodeInterpreter.ExecutionResult baseline = sandbox.codeInterpreter.runCode(
                "counter = 1\nprint(f'Initialized counter = {counter}')");
        System.out.print("[STDOUT] " + baseline.getStdout());

        sandbox.codeInterpreter.runCode(
                "counter += 1\nprint(f'Counter after second call = {counter}')",
                new RunCodeOptions()
                        .setOnStdout(msg -> System.out.print("[STDOUT] " + msg))
                        .setOnStderr(msg -> System.out.print("[STDERR] " + msg))
                        .setOnError(err -> System.out.println("[ERROR] " + err.getName() + ": " + err.getValue())));

        System.out.println("\n" + "=".repeat(60));
        System.out.println("Timeout Handling");
        System.out.println("=".repeat(60));

        try {
            sandbox.codeInterpreter.runCode(
                    "import time\nprint('Starting long running task...')\ntime.sleep(5)\nprint('Finished!')",
                    new RunCodeOptions()
                            .setTimeout(1)
                            .setOnStdout(msg -> System.out.print("[STDOUT] " + msg)));
        } catch (DaytonaTimeoutException e) {
            System.out.println("Timed out as expected: " + e.getMessage());
        }
    }
}
