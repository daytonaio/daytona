// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when the authenticated user lacks permission to perform an operation (HTTP 403).
 *
 * <pre>{@code
 * try {
 *     daytona.sandbox().delete(sandboxId);
 * } catch (DaytonaForbiddenException e) {
 *     System.err.println("Not authorized to delete this sandbox");
 * }
 * }</pre>
 */
public class DaytonaForbiddenException extends DaytonaException {
    /**
     * Creates a forbidden exception.
     *
     * @param message error description from the API
     */
    public DaytonaForbiddenException(String message) {
        super(403, message);
    }

    /**
     * @param message error description from the API
     * @param cause root cause
     */
    public DaytonaForbiddenException(String message, Throwable cause) {
        super(403, message, cause);
    }
}
