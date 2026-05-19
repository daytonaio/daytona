// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when a requested resource does not exist (HTTP 404).
 */
public class DaytonaNotFoundException extends DaytonaException {
    /**
     * Creates a not-found exception.
     *
     * @param message error description from the API
     */
    public DaytonaNotFoundException(String message) {
        super(404, message);
    }

    /**
     * @param message error description from the API
     * @param cause root cause
     */
    public DaytonaNotFoundException(String message, Throwable cause) {
        super(404, message, cause);
    }
}
