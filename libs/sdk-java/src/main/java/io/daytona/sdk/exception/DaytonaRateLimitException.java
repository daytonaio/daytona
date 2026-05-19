// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when API rate limits are exceeded (HTTP 429).
 */
public class DaytonaRateLimitException extends DaytonaException {
    /**
     * Creates a rate-limit exception.
     *
     * @param message error description from the API
     */
    public DaytonaRateLimitException(String message) {
        super(429, message);
    }

    /**
     * @param message error description from the API
     * @param cause root cause
     */
    public DaytonaRateLimitException(String message, Throwable cause) {
        super(429, message, cause);
    }
}
