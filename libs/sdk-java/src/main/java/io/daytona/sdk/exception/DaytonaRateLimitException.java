// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when API rate limits are exceeded (HTTP 429).
 */
public class DaytonaRateLimitException extends DaytonaException {
    /** HTTP status code carried by every instance of this class. */
    public static final int STATUS_CODE = 429;

    /**
     * Creates a rate-limit exception.
     *
     * @param message error description from the API
     */
    public DaytonaRateLimitException(String message) {
        super(STATUS_CODE, message);
    }

    /**
     * @param message error description from the API
     * @param cause root cause
     */
    public DaytonaRateLimitException(String message, Throwable cause) {
        super(STATUS_CODE, message, cause);
    }

    public DaytonaRateLimitException(String message, String code, String source) {
        super(STATUS_CODE, message, code, source);
    }

    public DaytonaRateLimitException(String message, Throwable cause, String code, String source) {
        super(STATUS_CODE, message, cause, code, source);
    }
}
