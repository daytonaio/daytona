// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The requested operation is not supported for this sandbox.
 *
 * <p>Subclass of {@link DaytonaBadRequestException}.
 */
public class DaytonaSandboxOperationNotSupportedException extends DaytonaBadRequestException {
    public DaytonaSandboxOperationNotSupportedException(String message) {
        super(message);
    }

    public DaytonaSandboxOperationNotSupportedException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaSandboxOperationNotSupportedException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaSandboxOperationNotSupportedException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
