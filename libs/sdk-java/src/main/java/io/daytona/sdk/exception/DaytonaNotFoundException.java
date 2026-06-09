// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when a requested resource does not exist (HTTP 404).
 */
public class DaytonaNotFoundException extends DaytonaException {
    /** HTTP status code carried by every instance of this class. */
    public static final int STATUS_CODE = 404;

    /**
     * Creates a not-found exception.
     *
     * @param message error description from the API
     */
    public DaytonaNotFoundException(String message) {
        super(STATUS_CODE, message);
    }

    /**
     * @param message error description from the API
     * @param cause root cause
     */
    public DaytonaNotFoundException(String message, Throwable cause) {
        super(STATUS_CODE, message, cause);
    }

    public DaytonaNotFoundException(String message, String code, String source) {
        super(STATUS_CODE, message, code, source);
    }

    public DaytonaNotFoundException(String message, Throwable cause, String code, String source) {
        super(STATUS_CODE, message, cause, code, source);
    }
}
