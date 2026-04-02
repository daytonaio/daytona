// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when the request is malformed or contains invalid parameters (HTTP 400).
 *
 * <pre>{@code
 * try {
 *     daytona.sandbox().create(params);
 * } catch (DaytonaBadRequestException e) {
 *     System.err.println("Invalid request parameters: " + e.getMessage());
 * }
 * }</pre>
 */
public class DaytonaBadRequestException extends DaytonaException {
    /**
     * Creates a bad-request exception.
     *
     * @param message error description from the API
     */
    public DaytonaBadRequestException(String message) {
        super(400, message);
    }
}
