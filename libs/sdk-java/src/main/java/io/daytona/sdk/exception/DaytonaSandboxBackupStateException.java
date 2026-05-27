// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The sandbox's backup is in an invalid state for this operation.
 *
 * <p>Subclass of {@link DaytonaBadRequestException}.
 */
public class DaytonaSandboxBackupStateException extends DaytonaBadRequestException {
    public DaytonaSandboxBackupStateException(String message) {
        super(message);
    }

    public DaytonaSandboxBackupStateException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaSandboxBackupStateException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaSandboxBackupStateException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
