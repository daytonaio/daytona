// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when API credentials are missing or invalid (HTTP 401).
 *
 * <pre>{@code
 * try {
 *     daytona.sandbox().create();
 * } catch (DaytonaAuthenticationException e) {
 *     System.err.println("Invalid or missing API key");
 * }
 * }</pre>
 */
public class DaytonaAuthenticationException extends DaytonaException {
    /**
     * Creates an authentication exception.
     *
     * @param message error description from the API
     */
    public DaytonaAuthenticationException(String message) {
        super(401, message);
    }
}
