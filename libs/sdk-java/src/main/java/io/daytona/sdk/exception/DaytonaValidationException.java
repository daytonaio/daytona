// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised for semantic validation failures (HTTP 422).
 *
 * <p>Raised when the request is well-formed but the values fail business logic
 * validation (e.g., unsupported resource class, invalid configuration).
 *
 * <pre>{@code
 * try {
 *     daytona.sandbox().create(params);
 * } catch (DaytonaValidationException e) {
 *     System.err.println("Validation failed: " + e.getMessage());
 * }
 * }</pre>
 */
public class DaytonaValidationException extends DaytonaException {
    /**
     * Creates a validation exception.
     *
     * @param message error description from the API
     */
    public DaytonaValidationException(String message) {
        super(422, message);
    }
}
