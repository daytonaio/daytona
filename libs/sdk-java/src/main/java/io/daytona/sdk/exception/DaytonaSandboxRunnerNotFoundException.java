// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The runner backing the sandbox could not be located.
 *
 * <p>Subclass of {@link DaytonaNotFoundException}.
 */
public class DaytonaSandboxRunnerNotFoundException extends DaytonaNotFoundException {
    public DaytonaSandboxRunnerNotFoundException(String message) {
        super(message);
    }

    public DaytonaSandboxRunnerNotFoundException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaSandboxRunnerNotFoundException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaSandboxRunnerNotFoundException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
