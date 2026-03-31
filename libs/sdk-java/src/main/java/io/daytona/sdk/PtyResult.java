// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

public class PtyResult {
    private final int exitCode;
    private final String error;

    public PtyResult(int exitCode, String error) {
        this.exitCode = exitCode;
        this.error = error;
    }

    public int getExitCode() {
        return exitCode;
    }

    public String getError() {
        return error;
    }
}
