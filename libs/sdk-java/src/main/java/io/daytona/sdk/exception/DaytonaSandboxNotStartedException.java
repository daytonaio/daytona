// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The targeted sandbox is not running.
 *
 * <p>Subclass of {@link DaytonaBadRequestException}.
 */
public class DaytonaSandboxNotStartedException extends DaytonaBadRequestException {
    public DaytonaSandboxNotStartedException(String message) {
        super(message);
    }

    public DaytonaSandboxNotStartedException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaSandboxNotStartedException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaSandboxNotStartedException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
