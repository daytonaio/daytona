// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

/**
 * Final outcome of a PTY session.
 *
 * <p>Contains exit status and an optional error/exit reason reported by the PTY backend.
 */
public class PtyResult {
    private final int exitCode;
    private final String error;

    /**
     * Creates a PTY result object.
     *
     * @param exitCode exit code returned by the PTY process; negative values indicate no exit code
     * @param error optional error or exit reason
     */
    public PtyResult(int exitCode, String error) {
        this.exitCode = exitCode;
        this.error = error;
    }

    /**
     * Returns the process exit code.
     *
     * @return PTY process exit code
     */
    public int getExitCode() {
        return exitCode;
    }

    /**
     * Returns the PTY error or exit reason when available.
     *
     * @return error message, or {@code null} when the session ended successfully
     */
    public String getError() {
        return error;
    }
}
