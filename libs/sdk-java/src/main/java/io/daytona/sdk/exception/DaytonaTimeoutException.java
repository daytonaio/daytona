// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when an SDK operation times out.
 *
 * <p>This exception is generated client-side and is not tied to a single HTTP status code.
 */
public class DaytonaTimeoutException extends DaytonaException {
    /**
     * Creates a timeout exception with a cause.
     *
     * @param message timeout description
     * @param cause root cause
     */
    public DaytonaTimeoutException(String message, Throwable cause) {
        super(message, cause);
    }

    /**
     * Creates a timeout exception.
     *
     * @param message timeout description
     */
    public DaytonaTimeoutException(String message) {
        super(message);
    }
}
