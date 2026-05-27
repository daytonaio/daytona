// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The sandbox is in the middle of a state change; retry shortly.
 *
 * <p>Subclass of {@link DaytonaConflictException}.
 */
public class DaytonaSandboxStateChangeInProgressException extends DaytonaConflictException {
    public DaytonaSandboxStateChangeInProgressException(String message) {
        super(message);
    }

    public DaytonaSandboxStateChangeInProgressException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaSandboxStateChangeInProgressException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaSandboxStateChangeInProgressException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
