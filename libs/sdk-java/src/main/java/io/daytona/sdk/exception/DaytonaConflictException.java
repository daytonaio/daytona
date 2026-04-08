// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when an operation conflicts with the current state (HTTP 409).
 *
 * <p>Common causes: creating a resource with a name that already exists,
 * or performing an operation incompatible with the resource's current state.
 *
 * <pre>{@code
 * try {
 *     daytona.snapshot().create(params);
 * } catch (DaytonaConflictException e) {
 *     System.err.println("A snapshot with this name already exists");
 * }
 * }</pre>
 */
public class DaytonaConflictException extends DaytonaException {
    /**
     * Creates a conflict exception.
     *
     * @param message error description from the API
     */
    public DaytonaConflictException(String message) {
        super(409, message);
    }
}
