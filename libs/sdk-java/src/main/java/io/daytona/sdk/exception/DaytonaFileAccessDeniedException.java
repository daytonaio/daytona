// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * Insufficient permissions for the filesystem operation.
 *
 * <p>Subclass of {@link DaytonaForbiddenException}.
 */
public class DaytonaFileAccessDeniedException extends DaytonaForbiddenException {
    public DaytonaFileAccessDeniedException(String message) {
        super(message);
    }

    public DaytonaFileAccessDeniedException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaFileAccessDeniedException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaFileAccessDeniedException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
