// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * A process exceeded its configured execution timeout.
 *
 * <p>Subclass of {@link DaytonaTimeoutException}.
 */
public class DaytonaProcessExecutionTimeoutException extends DaytonaTimeoutException {
    public DaytonaProcessExecutionTimeoutException(String message) {
        super(message);
    }

    public DaytonaProcessExecutionTimeoutException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaProcessExecutionTimeoutException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaProcessExecutionTimeoutException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
