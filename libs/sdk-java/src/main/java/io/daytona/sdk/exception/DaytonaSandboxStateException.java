// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The sandbox state does not allow this operation.
 *
 * <p>Subclass of {@link DaytonaBadRequestException}.
 */
public class DaytonaSandboxStateException extends DaytonaBadRequestException {
    public DaytonaSandboxStateException(String message) {
        super(message);
    }

    public DaytonaSandboxStateException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaSandboxStateException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaSandboxStateException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
