// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The sandbox disk expansion limit was reached.
 *
 * <p>Subclass of {@link DaytonaForbiddenException}.
 */
public class DaytonaSandboxDiskExpansionLimitException extends DaytonaForbiddenException {
    public DaytonaSandboxDiskExpansionLimitException(String message) {
        super(message);
    }

    public DaytonaSandboxDiskExpansionLimitException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaSandboxDiskExpansionLimitException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaSandboxDiskExpansionLimitException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
