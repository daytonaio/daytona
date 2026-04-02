// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.function.Consumer;

/**
 * Options for {@link CodeInterpreter#runCode(String, RunCodeOptions)}.
 */
public class RunCodeOptions {
    private Integer timeout;
    private Consumer<String> onStdout;
    private Consumer<String> onStderr;
    private Consumer<CodeInterpreter.ExecutionError> onError;

    public RunCodeOptions() {}

    public Integer getTimeout() { return timeout; }

    public RunCodeOptions setTimeout(Integer timeout) {
        this.timeout = timeout;
        return this;
    }

    public Consumer<String> getOnStdout() { return onStdout; }

    public RunCodeOptions setOnStdout(Consumer<String> onStdout) {
        this.onStdout = onStdout;
        return this;
    }

    public Consumer<String> getOnStderr() { return onStderr; }

    public RunCodeOptions setOnStderr(Consumer<String> onStderr) {
        this.onStderr = onStderr;
        return this;
    }

    public Consumer<CodeInterpreter.ExecutionError> getOnError() { return onError; }

    public RunCodeOptions setOnError(Consumer<CodeInterpreter.ExecutionError> onError) {
        this.onError = onError;
        return this;
    }
}
